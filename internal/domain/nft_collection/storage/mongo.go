package storage

import (
	"context"
	"fmt"
	"time"

	nftcollection "github.com/rom6n/create-nft-go/internal/domain/nft_collection"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type nftCollectionRepo struct {
	client         *mongo.Client
	dbName         string
	collectionName string
	timeout        time.Duration
}

type NftCollectionRepoCfg struct {
	DBName         string
	CollectionName string
	Timeout        time.Duration
}

func NewNftCollectionRepo(client *mongo.Client, cfg NftCollectionRepoCfg) nftcollection.NftCollectionRepository {
	return &nftCollectionRepo{
		client:         client,
		dbName:         cfg.DBName,
		collectionName: cfg.CollectionName,
		timeout:        cfg.Timeout,
	}
}

func (v *nftCollectionRepo) GetDatabase() *mongo.Database {
	return v.client.Database(v.dbName)
}

func (v *nftCollectionRepo) GetCollection() *mongo.Collection {
	return v.client.Database(v.dbName).Collection(v.collectionName)
}

func (v *nftCollectionRepo) GetContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, v.timeout)
}

func (v *nftCollectionRepo) CreateNftCollection(ctx context.Context, nftCollection *nftcollection.NftCollection) error {
	dbCtx, cancel := v.GetContext(ctx)
	defer cancel()

	collection := v.GetCollection()

	_, insertErr := collection.InsertOne(dbCtx, *nftCollection)
	if insertErr != nil {
		return insertErr
	}

	return nil
}

func (v *nftCollectionRepo) DeleteNftCollection(ctx context.Context, collectionAddress string) error {
	dbCtx, cancel := v.GetContext(ctx)
	defer cancel()

	collection := v.GetCollection()

	_, deleteErr := collection.DeleteOne(dbCtx, bson.D{{Key: "_id", Value: collectionAddress}})
	if deleteErr != nil {
		return deleteErr
	}

	return nil
}

func (v *nftCollectionRepo) GetNftCollectionByAddress(ctx context.Context, collectionAddress string) (*nftcollection.NftCollection, error) {
	dbCtx, cancel := v.GetContext(ctx)
	defer cancel()

	collection := v.GetCollection()

	var foundedCollection nftcollection.NftCollection
	decodeErr := collection.FindOne(dbCtx, bson.D{{Key: "_id", Value: collectionAddress}}).Decode(&foundedCollection)
	if decodeErr != nil {
		return &foundedCollection, fmt.Errorf("nft collection decode error after seaching: %v", decodeErr)
	}

	return &foundedCollection, nil
}
