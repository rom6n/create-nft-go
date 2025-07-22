package storage

import (
	"context"
	"time"

	"github.com/rom6n/create-nft-go/internal/domain/wallet"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type mongoWalletRepository struct {
	client         *mongo.Client
	dbName         string
	collectionName string
	timeout        time.Duration
}

type WalletRepositoryCfg struct {
	DBName           string
	WalletCollection string
	Timeout          time.Duration
}

func NewWalletRepository(client *mongo.Client, cfg WalletRepositoryCfg) wallet.WalletRepository {
	return newMongoWalletRepository(
		client,
		cfg.DBName,
		cfg.WalletCollection,
		cfg.Timeout,
	)
}

func newMongoWalletRepository(client *mongo.Client, dbName, collectionName string, timeout time.Duration) wallet.WalletRepository {
	return &mongoWalletRepository{
		client:         client,
		dbName:         dbName,
		collectionName: collectionName,
		timeout:        timeout,
	}
}

func (r *mongoWalletRepository) GetContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, r.timeout)
}

func (r *mongoWalletRepository) GetCollection() *mongo.Collection {
	return r.client.Database(r.dbName).Collection(r.collectionName)
}

func (r *mongoWalletRepository) AddWallet(ctx context.Context, wallet *wallet.Wallet) error {
	dbCtx, close := r.GetContext(ctx)
	defer close()

	walletsCollection := r.GetCollection()

	if _, insertErr := walletsCollection.InsertOne(dbCtx, *wallet); insertErr != nil {
		return insertErr
	}

	return nil
}

func (r *mongoWalletRepository) RefreshWalletNftItems(ctx context.Context, walletAddress string, nftItems *[]wallet.NftItem) error {
	dbCtx, close := r.GetContext(ctx)
	defer close()

	walletsCollection := r.GetCollection()

	filter := bson.D{{Key: "_id", Value: walletAddress}}

	update := bson.D{{Key: "$set", Value: bson.D{{Key: "nft_items", Value: *nftItems}}}}

	if _, updateErr := walletsCollection.UpdateOne(dbCtx, filter, update); updateErr != nil {
		return updateErr
	}

	return nil
}

func (r *mongoWalletRepository) GetWalletByAddress(ctx context.Context, address string) (*wallet.Wallet, error) {
	dbCtx, close := context.WithTimeout(ctx, 5*time.Second)
	defer close()

	walletsCollection := r.GetCollection()

	var wallet *wallet.Wallet

	if findErr := walletsCollection.FindOne(dbCtx, bson.D{{Key: "_id", Value: address}}).Decode(&wallet); findErr != nil {
		return nil, findErr
	}

	return wallet, nil
}
