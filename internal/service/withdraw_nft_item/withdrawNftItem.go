package withdrawnftitem

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"log"
	"time"

	nftitem "github.com/rom6n/create-nft-go/internal/domain/nft_item"
	"github.com/rom6n/create-nft-go/internal/domain/user"
	nftitemutils "github.com/rom6n/create-nft-go/internal/utils/contract_utils/nft_item_utils"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/nft"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

type WithdrawNftItemServiceRepository interface {
	WithdrawNftItem(ctx context.Context, nftItemAddress *address.Address, withdrawToAddress *address.Address, ownerID int64, isTestnet bool) error
}

type withdrawNftItemServiceRepo struct {
	nftItemRepo       nftitem.NftItemRepository
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

type WithdrawNftItemServiceCfg struct {
	NftItemRepo       nftitem.NftItemRepository
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

func New(cfg WithdrawNftItemServiceCfg) WithdrawNftItemServiceRepository {
	return &withdrawNftItemServiceRepo{
		cfg.NftItemRepo,
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

func (v *withdrawNftItemServiceRepo) getContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, v.timeout)
}

func (v *withdrawNftItemServiceRepo) WithdrawNftItem(ctx context.Context, nftItemAddress *address.Address, withdrawToAddress *address.Address, ownerID int64, isTestnet bool) error {
	svcCtx, cancel := v.getContext(ctx)
	defer cancel()

	nanoTonForWithdraw := uint64(30000000)

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

	nftItem, nftItemErr := v.nftItemRepo.GetNftItemByAddress(svcCtx, nftItemAddress.String())
	if nftItemErr != nil {
		return fmt.Errorf("error getting nft item: %v", nftItemErr)
	}

	if nftItem.Owner != ownerAccount.UUID {
		return fmt.Errorf("user must be an item's owner to withdraw it")
	}

	block, blockErr := api.GetMasterchainInfo(apiCtx)
	if blockErr != nil {
		return fmt.Errorf("error getting masterchain info: %v", blockErr)
	}

	nftItemClient := nft.NewItemClient(api, nftItemAddress)
	nftItemData, methodErr := nftItemClient.GetNFTDataAtBlock(apiCtx, block)
	if methodErr != nil {
		return fmt.Errorf("nft item get data method error: %v", methodErr)
	}

	if !walletAddress.Equals(nftItemData.OwnerAddress) {
		return fmt.Errorf("marketplace contract must be an nft item's owner to withdraw nft item")
	}

	changeOwnerMsg := nftitemutils.PackChangeOwnerMsg(withdrawToAddress, walletAddress, nftItemAddress)

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
