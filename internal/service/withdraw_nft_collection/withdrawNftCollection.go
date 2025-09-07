package withdrawnftcollection

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"log"
	"time"

	nftcollection "github.com/rom6n/create-nft-go/internal/domain/nft_collection"
	"github.com/rom6n/create-nft-go/internal/domain/user"
	nftcollectionutils "github.com/rom6n/create-nft-go/internal/utils/contract_utils/nft_collection_utils"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/nft"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

type WithdrawNftCollectionServiceRepository interface {
	WithdrawNftCollection(ctx context.Context, nftCollectionAddress *address.Address, withdrawToAddress *address.Address, ownerID int64, isTestnet bool) error
}

type withdrawNftCollectionServiceRepo struct {
	nftCollectionRepo nftcollection.NftCollectionRepository
	userRepo          user.UserRepository
	privateKey        ed25519.PrivateKey
	testnetLiteClient *liteclient.ConnectionPool
	mainnetLiteClient *liteclient.ConnectionPool
	testnetLiteApi    ton.APIClientWrapped
	mainnetLiteApi    ton.APIClientWrapped
	testnetWallet     *wallet.Wallet
	mainnetWallet     *wallet.Wallet
	timeout           time.Duration
}

type WithdrawNftCollectionServiceCfg struct {
	NftCollectionRepo nftcollection.NftCollectionRepository
	UserRepo          user.UserRepository
	PrivateKey        ed25519.PrivateKey
	TestnetLiteClient *liteclient.ConnectionPool
	MainnetLiteClient *liteclient.ConnectionPool
	MainnetLiteApi    ton.APIClientWrapped
	TestnetLiteApi    ton.APIClientWrapped
	TestnetWallet     *wallet.Wallet
	MainnetWallet     *wallet.Wallet
	Timeout           time.Duration
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
		cfg.TestnetWallet,
		cfg.MainnetWallet,
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
	walletAddress := v.testnetWallet.WalletAddress()
	w := v.testnetWallet
	if !isTestnet {
		client = v.mainnetLiteClient
		api = v.mainnetLiteApi
		walletAddress = v.mainnetWallet.WalletAddress()
		w = v.mainnetWallet
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

	block, blockErr := api.GetMasterchainInfo(apiCtx)
	if blockErr != nil {
		return fmt.Errorf("error getting masterchain info: %v", blockErr)
	}

	collectionClient := nft.NewCollectionClient(api, nftCollectionAddress)
	collectionData, methodErr := collectionClient.GetCollectionDataAtBlock(apiCtx, block)
	if methodErr != nil {
		return fmt.Errorf("nft collection get data method error: %v", methodErr)
	}

	// check if wallet is an owner of nft collection
	if !walletAddress.Equals(collectionData.OwnerAddress) {
		return fmt.Errorf("marketplace contract must be an nft collection's owner to withdraw nft collection")
	}

	changeOwnerMsg := nftcollectionutils.PackChangeOwnerMsg(withdrawToAddress, nftCollectionAddress)

	if updErr := v.userRepo.UpdateUserBalance(svcCtx, ownerAccount.UUID, ownerAccount.NanoTon-nanoTonForWithdraw); updErr != nil {
		return fmt.Errorf("error reducing user's balance: %v", updErr)
	}

	msg := &wallet.Message{
		Mode:            0,
		InternalMessage: changeOwnerMsg,
	}

	if msgErr := w.Send(apiCtx, msg, true); msgErr != nil {
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

	if delErr := v.nftCollectionRepo.DeleteNftCollection(svcCtx, nftCollectionAddress.String()); delErr != nil {
		log.Printf("error deleting nft collection from db after withdraw %v\n", delErr)
	}

	return nil
}
