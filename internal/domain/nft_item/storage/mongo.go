package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	nftitem "github.com/rom6n/create-nft-go/internal/domain/nft_item"
	"go.mongodb.org/mongo-driver/v2/bson"
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

func (v *nftItemRepo) GetClient() *mongo.Client {
	return v.client
}

func (v *nftItemRepo) getCollection() *mongo.Collection {
	return v.client.Database(v.dbName).Collection(v.collectionName)
}

func (v *nftItemRepo) getContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, v.timeout)
}

func (v *nftItemRepo) CreateNftItem(ctx context.Context, nftItem *nftitem.NftItem) error {
	dbCtx, cancel := v.getContext(ctx)
	defer cancel()

	collection := v.getCollection()

	_, insertErr := collection.InsertOne(dbCtx, *nftItem)
	if insertErr != nil {
		return insertErr
	}

	return nil
}

func (v *nftItemRepo) GetNftItemsByOwnerUuid(ctx context.Context, uuid uuid.UUID) ([]nftitem.NftItem, error) {
	dbCtx, cancel := v.getContext(ctx)
	defer cancel()

	collection := v.getCollection()

	var foundedCollections []nftitem.NftItem
	cursor, decodeErr := collection.Find(dbCtx, bson.D{{Key: "owner", Value: uuid}})
	if decodeErr != nil {
		return nil, fmt.Errorf("nft items decode error after seaching: %v", decodeErr)
	}

	if decodeErr2 := cursor.All(dbCtx, &foundedCollections); decodeErr2 != nil {
		return nil, fmt.Errorf("nft items decode error after find: %v", decodeErr2)
	}

	return foundedCollections, nil
}

func (v *nftItemRepo) GetNftItemByAddress(ctx context.Context, nftItemAddress string) (*nftitem.NftItem, error) {
	dbCtx, cancel := v.getContext(ctx)
	defer cancel()

	collection := v.getCollection()

	var foundedNftItem nftitem.NftItem
	decodeErr := collection.FindOne(dbCtx, bson.D{{Key: "_id", Value: nftItemAddress}}).Decode(&foundedNftItem)
	if decodeErr != nil {
		return &foundedNftItem, fmt.Errorf("nft item decode error after seaching: %v", decodeErr)
	}

	return &foundedNftItem, nil
}

func (v *nftItemRepo) DeleteNftItem(ctx context.Context, nftItemAddress string) error {
	dbCtx, cancel := v.getContext(ctx)
	defer cancel()

	collection := v.getCollection()

	if _, err := collection.DeleteOne(dbCtx, bson.D{{Key: "_id", Value: nftItemAddress}}); err != nil {
		return fmt.Errorf("error deleting nft item from db: %v", err)
	}

	return nil
}
