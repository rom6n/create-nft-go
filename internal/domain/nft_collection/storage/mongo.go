package storage

import (
	"context"
	"time"

	"github.com/google/uuid"
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

func (v *nftCollectionRepo) CreateCollection(ctx context.Context, nftCollection *nftcollection.NftCollection, ownerUuid uuid.UUID) error {
	dbCtx, cancel := context.WithTimeout(ctx, v.timeout)
	defer cancel()

	collection := v.GetCollection()

	_, insertErr := collection.InsertOne(dbCtx, *nftCollection)
	if insertErr != nil {
		return insertErr
	}

	return nil
}

func (v *nftCollectionRepo) DeleteCollection(ctx context.Context, collectionAddress string) error {
	dbCtx, cancel := context.WithTimeout(ctx, v.timeout)
	defer cancel()

	collection := v.GetCollection()

	_, deleteErr := collection.DeleteOne(dbCtx, bson.D{{Key: "_id", Value: collectionAddress}})
	if deleteErr != nil {
		return deleteErr
	}

	return nil
}

