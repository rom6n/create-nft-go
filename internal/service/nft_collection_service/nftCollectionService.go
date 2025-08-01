package nftcollectionservice

import (
	"context"
	"crypto/ed25519"
	"log"
	"time"

	nftcollection "github.com/rom6n/create-nft-go/internal/domain/nft_collection"
	userservice "github.com/rom6n/create-nft-go/internal/service/user_service"
	marketutils "github.com/rom6n/create-nft-go/internal/utils/contract_utils/market_utils"
	"github.com/rom6n/create-nft-go/internal/utils/tonutil"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type NftCollectionServiceRepository interface {
	DeployMarketplaceContract(ctx context.Context) error
}

type nftCollectionServiceRepo struct {
	NftCollectionRepo       nftcollection.NftCollectionRepository
	UserRepo                userservice.UserServiceRepository
	PrivateKey              ed25519.PrivateKey
	LiteClient              *liteclient.ConnectionPool
	LiteclientApi           ton.APIClientWrapped
	MarketplaceContractCode *cell.Cell
}

type NftCollectionServiceCfg struct {
	NftCollectionRepo       nftcollection.NftCollectionRepository
	UserRepo                userservice.UserServiceRepository
	PrivateKey              ed25519.PrivateKey
	LiteClient              *liteclient.ConnectionPool
	LiteclientApi           ton.APIClientWrapped
	MarketplaceContractCode *cell.Cell
}

func New(cfg NftCollectionServiceCfg) NftCollectionServiceRepository {
	return &nftCollectionServiceRepo{
		cfg.NftCollectionRepo,
		cfg.UserRepo,
		cfg.PrivateKey,
		cfg.LiteClient,
		cfg.LiteclientApi,
		cfg.MarketplaceContractCode,
	}
}

func (v *nftCollectionServiceRepo) DeployMarketplaceContract(ctx context.Context) error {
	svcCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	api := v.LiteclientApi

	walletV4, walletErr := tonutil.GetTestWallet(api)
	if walletErr != nil {
		return walletErr
	}

	msgBody := cell.BeginCell().EndCell()
	deployedAddr, _, _, deployErr := walletV4.DeployContractWaitTransaction(
		svcCtx,
		tlb.MustFromTON("0.07"),
		msgBody,
		v.MarketplaceContractCode,
		marketutils.GetMarketplaceContractDeployData(0, 1947320581, []byte(v.PrivateKey.Public().(ed25519.PublicKey))),
	)
	if deployErr != nil {
		log.Printf("Error deploy market contract: %v\n", deployErr)
		return deployErr
	}

	log.Printf("Market contract deployed at address: %v\n", deployedAddr)

	return nil
}
