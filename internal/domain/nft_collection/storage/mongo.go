package storage

import (
	"context"
	"fmt"
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

func (v *nftCollectionRepo) GetClient() *mongo.Client {
	return v.client
}

func (v *nftCollectionRepo) getCollection() *mongo.Collection {
	return v.client.Database(v.dbName).Collection(v.collectionName)
}

func (v *nftCollectionRepo) getContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, v.timeout)
}

func (v *nftCollectionRepo) CreateNftCollection(ctx context.Context, nftCollection *nftcollection.NftCollection) error {
	dbCtx, cancel := v.getContext(ctx)
	defer cancel()

	collection := v.getCollection()

	_, insertErr := collection.InsertOne(dbCtx, *nftCollection)
	if insertErr != nil {
		return insertErr
	}

	return nil
}

func (v *nftCollectionRepo) DeleteNftCollection(ctx context.Context, collectionAddress string) error {
	dbCtx, cancel := v.getContext(ctx)
	defer cancel()

	collection := v.getCollection()

	_, deleteErr := collection.DeleteOne(dbCtx, bson.D{{Key: "_id", Value: collectionAddress}})
	if deleteErr != nil {
		return deleteErr
	}

	return nil
}

func (v *nftCollectionRepo) GetNftCollectionByAddress(ctx context.Context, collectionAddress string) (*nftcollection.NftCollection, error) {
	dbCtx, cancel := v.getContext(ctx)
	defer cancel()

	collection := v.getCollection()

	var foundedCollection nftcollection.NftCollection
	decodeErr := collection.FindOne(dbCtx, bson.D{{Key: "_id", Value: collectionAddress}}).Decode(&foundedCollection)
	if decodeErr != nil {
		return &foundedCollection, fmt.Errorf("nft collection decode error after seaching: %v", decodeErr)
	}

	return &foundedCollection, nil
}

func (v *nftCollectionRepo) GetNftCollectionsByOwnerUuid(ctx context.Context, uuid uuid.UUID) ([]nftcollection.NftCollection, error) {
	dbCtx, cancel := v.getContext(ctx)
	defer cancel()

	collection := v.getCollection()

	var foundedCollections []nftcollection.NftCollection
	cursor, decodeErr := collection.Find(dbCtx, bson.D{{Key: "owner", Value: uuid}})
	if decodeErr != nil {
		return nil, fmt.Errorf("nft collections decode error after seaching: %v", decodeErr)
	}

	if decodeErr2 := cursor.All(dbCtx, &foundedCollections); decodeErr2 != nil {
		return nil, fmt.Errorf("nft collections decode error after find: %v", decodeErr2)
	}

	return foundedCollections, nil
}
