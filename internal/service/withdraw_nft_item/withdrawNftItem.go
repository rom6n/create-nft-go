package withdrawnftitem

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"log"
	"time"

	nftitem "github.com/rom6n/create-nft-go/internal/domain/nft_item"
	"github.com/rom6n/create-nft-go/internal/domain/user"
	marketutils "github.com/rom6n/create-nft-go/internal/utils/contract_utils/market_utils"
	nftitemutils "github.com/rom6n/create-nft-go/internal/utils/contract_utils/nft_item_utils"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/nft"
)

type WithdrawNftItemServiceRepository interface {
	WithdrawNftItem(ctx context.Context, nftItemAddress *address.Address, withdrawToAddress *address.Address, ownerID int64, isTestnet bool) error
}

type withdrawNftItemServiceRepo struct {
	nftItemRepo                       nftitem.NftItemRepository
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

type WithdrawNftItemServiceCfg struct {
	NftItemRepo                       nftitem.NftItemRepository
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

func New(cfg WithdrawNftItemServiceCfg) WithdrawNftItemServiceRepository {
	return &withdrawNftItemServiceRepo{
		cfg.NftItemRepo,
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

func (v *withdrawNftItemServiceRepo) getContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, v.timeout)
}

func (v *withdrawNftItemServiceRepo) WithdrawNftItem(ctx context.Context, nftItemAddress *address.Address, withdrawToAddress *address.Address, ownerID int64, isTestnet bool) error {
	svcCtx, cancel := v.getContext(ctx)
	defer cancel()

	nanoTonForWithdraw := uint64(50000000)

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

	nftItem, nftItemErr := v.nftItemRepo.GetNftItemByAddress(svcCtx, nftItemAddress.String())
	if nftItemErr != nil {
		return fmt.Errorf("error getting nft item: %v", nftItemErr)
	}

	if nftItem.Owner != ownerAccount.UUID {
		return fmt.Errorf("user must be an item's owner to withdraw it")
	}

	nftItemClient := nft.NewItemClient(api, nftItemAddress)
	nftItemData, methodErr := nftItemClient.GetNFTData(apiCtx)
	if methodErr != nil {
		return fmt.Errorf("nft item get data method error: %v", methodErr)
	}

	if !marketplaceContractAddress.Equals(nftItemData.OwnerAddress) {
		return fmt.Errorf("marketplace contract must be an nft item's owner to withdraw nft item")
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

	msgToSend := nftitemutils.PackChangeOwnerMsg(withdrawToAddress, marketplaceContractAddress, nftItemAddress)

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
				return fmt.Errorf("error returning ton to user & error sending withdraw nft item external message: %v", msgErr)
			}
			time.Sleep(1 * time.Second)
		}
		return fmt.Errorf("error sending external message to withdraw nft item: %v", msgErr)
	}

	if delErr := v.nftItemRepo.DeleteNftItem(svcCtx, nftItemAddress.String()); delErr != nil {
		log.Printf("error deleting nft item from db after withdraw %v\n", delErr)
	}

	return nil
}
