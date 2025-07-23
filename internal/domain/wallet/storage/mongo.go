package storage

import (
	"context"
	"time"

	"github.com/rom6n/create-nft-go/internal/domain/wallet"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type mongoWalletRepo struct {
	client         *mongo.Client
	dbName         string
	collectionName string
	timeout        time.Duration
}

type WalletRepoCfg struct {
	DBName         string
	CollectionName string
	Timeout        time.Duration
}

func NewWalletRepo(client *mongo.Client, cfg WalletRepoCfg) wallet.WalletRepository {
	return &mongoWalletRepo{
		client:         client,
		dbName:         cfg.DBName,
		collectionName: cfg.CollectionName,
		timeout:        cfg.Timeout,
	}
}

func (r *mongoWalletRepo) GetContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, r.timeout)
}

func (r *mongoWalletRepo) GetCollection() *mongo.Collection {
	return r.client.Database(r.dbName).Collection(r.collectionName)
}

func (r *mongoWalletRepo) AddWallet(ctx context.Context, wallet *wallet.Wallet) error {
	dbCtx, close := r.GetContext(ctx)
	defer close()

	walletsCollection := r.GetCollection()

	if _, insertErr := walletsCollection.InsertOne(dbCtx, *wallet); insertErr != nil {
		return insertErr
	}

	return nil
}

func (r *mongoWalletRepo) UpdateWalletNftItems(ctx context.Context, walletAddress string, nftItems *[]wallet.NftItem) error {
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

func (r *mongoWalletRepo) GetWalletByAddress(ctx context.Context, address string) (*wallet.Wallet, error) {
	dbCtx, close := context.WithTimeout(ctx, 5*time.Second)
	defer close()

	walletsCollection := r.GetCollection()

	var wallet *wallet.Wallet

	if findErr := walletsCollection.FindOne(dbCtx, bson.D{{Key: "_id", Value: address}}).Decode(&wallet); findErr != nil {
		return nil, findErr
	}

	return wallet, nil
}
