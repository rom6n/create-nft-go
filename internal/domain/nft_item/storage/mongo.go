package storage

import (
	"context"
	"time"

	nftitem "github.com/rom6n/create-nft-go/internal/domain/nft_item"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type nftItemRepo struct {
	client         *mongo.Client
	dbName         string
	collectionName string
	timeout        time.Duration
}

type NftItemRepoCfg struct {
	DBName         string
	CollectionName string
	Timeout        time.Duration
}

func NewNftItemRepo(client *mongo.Client, cfg NftItemRepoCfg) nftitem.NftItemRepository {
	return &nftItemRepo{
		client:         client,
		dbName:         cfg.DBName,
		collectionName: cfg.CollectionName,
		timeout:        cfg.Timeout,
	}
}

func (v *nftItemRepo) GetDatabase() *mongo.Database {
	return v.client.Database(v.dbName)
}

func (v *nftItemRepo) GetCollection() *mongo.Collection {
	return v.client.Database(v.dbName).Collection(v.collectionName)
}

func (v *nftItemRepo) GetContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, v.timeout)
}

func (v *nftItemRepo) CreateNftItem(ctx context.Context, nftItem *nftitem.NftItem) error {
	dbCtx, cancel := v.GetContext(ctx)
	defer cancel()

	collection := v.GetCollection()

	_, insertErr := collection.InsertOne(dbCtx, *nftItem)
	if insertErr != nil {
		return insertErr
	}

	return nil
}
