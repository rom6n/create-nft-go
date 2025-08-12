package marketplacecontractservice

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"log"
	"time"

	generalcontractutils "github.com/rom6n/create-nft-go/internal/utils/contract_utils/general_contract_utils"
	marketutils "github.com/rom6n/create-nft-go/internal/utils/contract_utils/market_utils"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type MarketplaceContractServiceRepository interface {
	DepositMarketplaceContract(ctx context.Context, amount uint64, isTestnet bool) error
	DeployMarketplaceContract(ctx context.Context, isTestnet bool, subwallet ...int32) error
	WithdrawTonFromMarketplaceContract(ctx context.Context, amount uint64, isTestnet bool, textMessage ...string) error
}

type marketplaceContractServiceRepo struct {
	testnetLiteClient                 *liteclient.ConnectionPool
	mainnetLiteClient                 *liteclient.ConnectionPool
	testnetLiteApi                    ton.APIClientWrapped
	mainnetLiteApi                    ton.APIClientWrapped
	testnetMarketplaceContractAddress *address.Address
	mainnetMarketplaceContractAddress *address.Address
	testnetWallet                     *wallet.Wallet
	mainnetWallet                     *wallet.Wallet
	privateKey                        ed25519.PrivateKey
	marketplaceContractCode           *cell.Cell
	timeout                           time.Duration
}

type MarketplaceContractServiceCfg struct {
	TestnetLiteClient                 *liteclient.ConnectionPool
	MainnetLiteClient                 *liteclient.ConnectionPool
	TestnetLiteApi                    ton.APIClientWrapped
	MainnetLiteApi                    ton.APIClientWrapped
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
		testnetLiteClient:                 cfg.TestnetLiteClient,
		mainnetLiteClient:                 cfg.MainnetLiteClient,
		testnetLiteApi:                    cfg.TestnetLiteApi,
		mainnetLiteApi:                    cfg.MainnetLiteApi,
		testnetMarketplaceContractAddress: cfg.TestnetMarketplaceContractAddress,
		mainnetMarketplaceContractAddress: cfg.MainnetMarketplaceContractAddress,
		testnetWallet:                     cfg.TestnetWallet,
		mainnetWallet:                     cfg.MainnetWallet,
		privateKey:                        cfg.PrivateKey,
		marketplaceContractCode:           cfg.MarketplaceContractCode,
		timeout:                           cfg.Timeout,
	}
}

func (v *marketplaceContractServiceRepo) getContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, v.timeout)
}

func (v *marketplaceContractServiceRepo) DepositMarketplaceContract(ctx context.Context, amount uint64, isTestnet bool) error {
	svcCtx, cancel := v.getContext(ctx)
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
		log.Printf("Error deploying market contract: %v\n", deployErr)
		return deployErr
	}

	log.Printf("Market contract deployed at address: %v\n", deployedAddr)

	return nil
}

func (v *marketplaceContractServiceRepo) WithdrawTonFromMarketplaceContract(ctx context.Context, amount uint64, isTestnet bool, textMessage ...string) error {
	svcCtx, cancel := v.getContext(ctx)
	defer cancel()

	client := v.testnetLiteClient
	api := v.testnetLiteApi
	marketplaceContractAddress := v.testnetMarketplaceContractAddress
	walletAddress := v.testnetWallet.Address()
	if !isTestnet {
		client = v.mainnetLiteClient
		api = v.testnetLiteApi
		marketplaceContractAddress = v.mainnetMarketplaceContractAddress
		walletAddress = v.mainnetWallet.Address()
	}

	if amount < 5000000 {
		return fmt.Errorf("minimal withdrawal amount is 0.005 TON (5000000 nanoTON)")
	}

	apiCtx := client.StickyContext(svcCtx)

	block, bErr := api.GetMasterchainInfo(apiCtx)
	if bErr != nil {
		return fmt.Errorf("error getting masterchain info: %v", bErr)
	}

	response, responseErr := api.WaitForBlock(block.SeqNo).RunGetMethod(apiCtx, block, marketplaceContractAddress, "seqno")
	if responseErr != nil {
		return fmt.Errorf("error getting marketplace contract seqno: %v", responseErr)
	}

	seqno, seqnoErr := response.Int(0)
	if seqnoErr != nil {
		return fmt.Errorf("marketplace method seqno returned not a seqno: %v", seqnoErr)
	}

	msgToSend := generalcontractutils.PackDefaultMessage(walletAddress, amount, textMessage...)

	merketplaceContractMsg := marketutils.PackMessageToMarketplaceContract(v.privateKey, time.Now().Add(1*time.Minute).Unix(), seqno, 0, msgToSend)

	msg := &tlb.ExternalMessage{
		DstAddr: marketplaceContractAddress,
		Body:    merketplaceContractMsg,
	}

	msgErr := api.SendExternalMessage(svcCtx, msg)
	if msgErr != nil {
		return fmt.Errorf("fail to send external message: %v", msgErr)
	}

	return nil
}
