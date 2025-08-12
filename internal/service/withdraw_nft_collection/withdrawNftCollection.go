package withdrawnftcollection

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"log"
	"time"

	nftcollection "github.com/rom6n/create-nft-go/internal/domain/nft_collection"
	"github.com/rom6n/create-nft-go/internal/domain/user"
	marketutils "github.com/rom6n/create-nft-go/internal/utils/contract_utils/market_utils"
	nftcollectionutils "github.com/rom6n/create-nft-go/internal/utils/contract_utils/nft_collection_utils"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/nft"
)

type WithdrawNftCollectionServiceRepository interface {
	WithdrawNftCollection(ctx context.Context, nftCollectionAddress *address.Address, withdrawToAddress *address.Address, ownerID int64, isTestnet bool) error
}

type withdrawNftCollectionServiceRepo struct {
	nftCollectionRepo                 nftcollection.NftCollectionRepository
	userRepo                          user.UserRepository
	privateKey                        ed25519.PrivateKey
	testnetLiteClient                 *liteclient.ConnectionPool
	mainnetLiteClient                 *liteclient.ConnectionPool
	testnetLiteApi                    ton.APIClientWrapped
	mainnetLiteApi                    ton.APIClientWrapped
	testnetMarketplaceContractAddress *address.Address
	mainnetMarketplaceContractAddress *address.Address
	timeout                           time.Duration
}

type WithdrawNftCollectionServiceCfg struct {
	NftCollectionRepo                 nftcollection.NftCollectionRepository
	UserRepo                          user.UserRepository
	PrivateKey                        ed25519.PrivateKey
	TestnetLiteClient                 *liteclient.ConnectionPool
	MainnetLiteClient                 *liteclient.ConnectionPool
	MainnetLiteApi                    ton.APIClientWrapped
	TestnetLiteApi                    ton.APIClientWrapped
	TestnetMarketplaceContractAddress *address.Address
	MainnetMarketplaceContractAddress *address.Address
	Timeout                           time.Duration
}

func New(cfg WithdrawNftCollectionServiceCfg) WithdrawNftCollectionServiceRepository {
	return &withdrawNftCollectionServiceRepo{
		cfg.NftCollectionRepo,
		cfg.UserRepo,
		cfg.PrivateKey,
		cfg.TestnetLiteClient,
		cfg.MainnetLiteClient,
		cfg.TestnetLiteApi,
		cfg.MainnetLiteApi,
		cfg.TestnetMarketplaceContractAddress,
		cfg.MainnetMarketplaceContractAddress,
		cfg.Timeout,
	}
}

func (v *withdrawNftCollectionServiceRepo) getContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, v.timeout)
}

func (v *withdrawNftCollectionServiceRepo) WithdrawNftCollection(ctx context.Context, nftCollectionAddress *address.Address, withdrawToAddress *address.Address, ownerID int64, isTestnet bool) error {
	svcCtx, cancel := v.getContext(ctx)
	defer cancel()

	nanoTonForWithdraw := uint64(10000000)

	client := v.testnetLiteClient
	api := v.testnetLiteApi
	marketplaceContractAddress := v.testnetMarketplaceContractAddress
	if !isTestnet {
		client = v.mainnetLiteClient
		api = v.mainnetLiteApi
		marketplaceContractAddress = v.mainnetMarketplaceContractAddress
	}

	apiCtx := client.StickyContext(svcCtx)

	ownerAccount, accErr := v.userRepo.GetUserByID(svcCtx, ownerID)
	if accErr != nil {
		return fmt.Errorf("error getting user's account: %v", accErr)
	}

	if ownerAccount.NanoTon < nanoTonForWithdraw {
		return fmt.Errorf("not enough ton, need %v more", nanoTonForWithdraw-ownerAccount.NanoTon)
	}

	nftCollection, collectionErr := v.nftCollectionRepo.GetNftCollectionByAddress(svcCtx, nftCollectionAddress.String())
	if collectionErr != nil {
		return fmt.Errorf("error getting nft collection: %v", collectionErr)
	}

	if nftCollection.Owner != ownerAccount.UUID {
		return fmt.Errorf("user must be an collection's owner to withdraw it")
	}

	collectionClient := nft.NewCollectionClient(api, nftCollectionAddress)
	collectionData, methodErr := collectionClient.GetCollectionData(apiCtx)
	if methodErr != nil {
		return fmt.Errorf("nft collection get data method error: %v", methodErr)
	}

	if !marketplaceContractAddress.Equals(collectionData.OwnerAddress) {
		return fmt.Errorf("marketplace contract must be an nft collection's owner to withdraw nft collection")
	}

	block, _ := api.GetMasterchainInfo(apiCtx)

	result, methodErr := api.WaitForBlock(block.SeqNo).RunGetMethod(apiCtx, block, marketplaceContractAddress, "seqno")
	if methodErr != nil {
		return fmt.Errorf("marketplace contract seqno method error: %v", methodErr)
	}

	seqno, seqnoErr := result.Int(0)
	if seqnoErr != nil {
		return fmt.Errorf("marketplace contract returned not a seqno")
	}

	msgToSend := nftcollectionutils.PackChangeOwnerMsg(withdrawToAddress, nftCollectionAddress)

	msgToMarketplace := marketutils.PackMessageToMarketplaceContract(v.privateKey, time.Now().Add(1*time.Minute).Unix(), seqno, 0, msgToSend)

	if updErr := v.userRepo.UpdateUserBalance(svcCtx, ownerAccount.UUID, ownerAccount.NanoTon-nanoTonForWithdraw); updErr != nil {
		return fmt.Errorf("error reducing user's balance: %v", updErr)
	}

	msg := &tlb.ExternalMessage{
		DstAddr: marketplaceContractAddress,
		Body:    msgToMarketplace,
	}

	if msgErr := api.SendExternalMessage(apiCtx, msg); msgErr != nil {
		for i := 0; i < 10; i++ {
			if updErr := v.userRepo.UpdateUserBalance(svcCtx, ownerAccount.UUID, ownerAccount.NanoTon); updErr == nil {
				break
			}
			log.Printf("Error returning %v ton to user, try: %v\n", nanoTonForWithdraw, i)
			if i == 9 {
				return fmt.Errorf("error returning ton to user & error sending withdraw nft collection external message: %v", msgErr)
			}
			time.Sleep(1 * time.Second)
		}
		return fmt.Errorf("error sending external message to withdraw nft collection: %v", msgErr)
	}

	return nil
}
