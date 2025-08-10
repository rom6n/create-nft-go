package nftcollectionservice

import (
	"crypto/ed25519"

	nftcollection "github.com/rom6n/create-nft-go/internal/domain/nft_collection"
	"github.com/rom6n/create-nft-go/internal/domain/user"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type NftCollectionServiceRepository interface {
}

type nftCollectionServiceRepo struct {
	NftCollectionRepo       nftcollection.NftCollectionRepository
	UserRepo                user.UserRepository
	PrivateKey              ed25519.PrivateKey
	LiteClient              *liteclient.ConnectionPool
	LiteclientApi           ton.APIClientWrapped
	MarketplaceContractCode *cell.Cell
}

type NftCollectionServiceCfg struct {
	NftCollectionRepo       nftcollection.NftCollectionRepository
	UserRepo                user.UserRepository
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
