package marketplacecontractservice

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"log"
	"time"

	marketutils "github.com/rom6n/create-nft-go/internal/utils/contract_utils/market_utils"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type MarketplaceContractServiceRepository interface {
	DepositMarketplaceContract(ctx context.Context, amount uint64, isTestnet bool) error
	DeployMarketplaceContract(ctx context.Context, isTestnet bool, subwallet ...int32) error
}

type marketplaceContractServiceRepo struct {
	testnetLiteClient *liteclient.ConnectionPool
	mainnetLiteClient *liteclient.ConnectionPool
	//testnetLiteApi                    ton.APIClientWrapped
	//mainnetLiteApi                    ton.APIClientWrapped
	testnetMarketplaceContractAddress *address.Address
	mainnetMarketplaceContractAddress *address.Address
	testnetWallet                     *wallet.Wallet
	mainnetWallet                     *wallet.Wallet
	privateKey                        ed25519.PrivateKey
	marketplaceContractCode           *cell.Cell
	timeout                           time.Duration
}

type MarketplaceContractServiceCfg struct {
	TestnetLiteClient *liteclient.ConnectionPool
	MainnetLiteClient *liteclient.ConnectionPool
	//TestnetLiteApi                    ton.APIClientWrapped
	//MainnetLiteApi                    ton.APIClientWrapped
	TestnetMarketplaceContractAddress *address.Address
	MainnetMarketplaceContractAddress *address.Address
	TestnetWallet                     *wallet.Wallet
	MainnetWallet                     *wallet.Wallet
	PrivateKey                        ed25519.PrivateKey
	MarketplaceContractCode           *cell.Cell
	Timeout                           time.Duration
}

func New(cfg MarketplaceContractServiceCfg) MarketplaceContractServiceRepository {
	return &marketplaceContractServiceRepo{
		testnetLiteClient: cfg.TestnetLiteClient,
		mainnetLiteClient: cfg.MainnetLiteClient,
		//testnetLiteApi:                    cfg.TestnetLiteApi,
		//mainnetLiteApi:                    cfg.MainnetLiteApi,
		testnetMarketplaceContractAddress: cfg.TestnetMarketplaceContractAddress,
		mainnetMarketplaceContractAddress: cfg.MainnetMarketplaceContractAddress,
		testnetWallet:                     cfg.TestnetWallet,
		mainnetWallet:                     cfg.MainnetWallet,
		privateKey:                        cfg.PrivateKey,
		marketplaceContractCode:           cfg.MarketplaceContractCode,
		timeout:                           cfg.Timeout,
	}
}

func (v *marketplaceContractServiceRepo) GetContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, v.timeout)
}

func (v *marketplaceContractServiceRepo) DepositMarketplaceContract(ctx context.Context, amount uint64, isTestnet bool) error {
	svcCtx, cancel := v.GetContext(ctx)
	defer cancel()

	client := v.testnetLiteClient
	marketplaceContractAddress := v.testnetMarketplaceContractAddress
	walletApi := v.testnetWallet
	if !isTestnet {
		client = v.mainnetLiteClient
		marketplaceContractAddress = v.mainnetMarketplaceContractAddress
		walletApi = v.mainnetWallet
	}

	apiCtx := client.StickyContext(svcCtx)

	if err := walletApi.Transfer(apiCtx, marketplaceContractAddress, tlb.FromNanoTONU(amount), fmt.Sprintf("Deposit from dev %v TON", tlb.FromNanoTONU(amount)), true); err != nil {
		return fmt.Errorf("failed to deposit: %v", err)
	}

	return nil
}

func (v *marketplaceContractServiceRepo) DeployMarketplaceContract(ctx context.Context, isTestnet bool, subwallet ...int32) error {
	svcCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	walletApi := v.testnetWallet
	if !isTestnet {
		walletApi = v.mainnetWallet
	}

	subw := int32(1947320581)
	if subwallet != nil {
		subw = subwallet[0]
	}

	msgBody := cell.BeginCell().EndCell()
	deployedAddr, _, _, deployErr := walletApi.DeployContractWaitTransaction(
		svcCtx,
		tlb.MustFromTON("0.05"),
		msgBody,
		v.marketplaceContractCode,
		marketutils.GetMarketplaceContractDeployData(0, subw, []byte(v.privateKey.Public().(ed25519.PublicKey))),
	)
	if deployErr != nil {
		log.Printf("Error deploy market contract: %v\n", deployErr)
		return deployErr
	}

	log.Printf("Market contract deployed at address: %v\n", deployedAddr)

	return nil
}
