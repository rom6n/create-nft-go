package nosql

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Attribute struct {
	TraitType string `bson:"trait_type" json:"trait_type"`
	Value     string `bson:"value" json:"value"`
}

type NftItemMetadata struct {
	Name        string      `bson:"name" json:"name"`
	Image       string      `bson:"image" json:"image"`
	Attributes  []Attribute `bson:"attributes" json:"attributes"`
	Description string      `bson:"description" json:"description"`
	ExternalUrl string      `bson:"external_url" json:"external_url"`
}

type NftItem struct {
	Address           string            `bson:"address" json:"address"`
	Index             int64             `bson:"index" json:"index"`
	CollectionAddress string            `bson:"collection_address" json:"collection_address"`
	CollectionName    string            `bson:"collection_name" json:"collection_name"`
	Owner             string            `bson:"owner" json:"owner"`
	Metadata          map[string]string `bson:"metadata" json:"metadata"` // под вопросом как метадата будет приходить
}

type NftCollection struct {
	Address        string            `bson:"address" json:"address"`
	NextItemIndex  int64             `bson:"next_item_index" json:"next_item_index"`
	Owner          string            `bson:"owner" json:"owner"`
	RoyaltyProcent int32             `bson:"royalty_procent" json:"royalty_procent"`
	Metadata       map[string]string `bson:"metadata" json:"metadata"` // под вопросом как метадата будет приходить
}

type Wallet struct {
	Address        string          `bson:"_id" json:"address"`
	NftCollections []NftCollection `bson:"nft_collections" json:"nft_collections"`
	NftItems       []NftItem       `bson:"nft_items" json:"nft_items"`
}

func NewMongoClient() *mongo.Client {
	if loadErr := godotenv.Load(); loadErr != nil {
		log.Fatal("Error load env.")
	}

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("Error. Add MongoDB uri in env.")
	}

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal("Error connect to MongoDB")
	}

	return client
}

func AddWalletToMongo(ctx context.Context, mongoClient *mongo.Client, wallet Wallet) error {
	dbCtx, close := context.WithTimeout(ctx, 10*time.Second)
	defer close()

	walletsCollection := mongoClient.Database("create_nft_tma").Collection("wallets")

	if _, insertErr := walletsCollection.InsertOne(dbCtx, wallet); insertErr != nil {
		return insertErr
	}

	return nil
}

func UpdateWalletInMongo(ctx context.Context, mongoClient *mongo.Client, wallet Wallet) error {
	dbCtx, close := context.WithTimeout(ctx, 10*time.Second)
	defer close()

	walletsCollection := mongoClient.Database("create_nft_tma").Collection("wallets")

	filter := bson.D{{Key: "_id", Value: wallet.Address}}

	update := bson.D{{Key: "$set", Value: bson.D{{Key: "nft_collections", Value: wallet.NftCollections}, {Key: "nft_items", Value: wallet.NftItems}}}}

	if _, updateErr := walletsCollection.UpdateOne(dbCtx, filter, update); updateErr != nil {
		return updateErr
	}

	return nil
}

func UpdateWalletNftItemsInMongo(ctx context.Context, mongoClient *mongo.Client, walletAddress string, nftItems []NftItem) error {
	dbCtx, close := context.WithTimeout(ctx, 10*time.Second)
	defer close()

	walletsCollection := mongoClient.Database("create_nft_tma").Collection("wallets")

	filter := bson.D{{Key: "_id", Value: walletAddress}}

	update := bson.D{{Key: "$set", Value: bson.D{{Key: "nft_items", Value: nftItems}}}}

	if _, updateErr := walletsCollection.UpdateOne(dbCtx, filter, update); updateErr != nil {
		return updateErr
	}

	return nil
}

func UpdateWalletNftCollectionsInMongo(ctx context.Context, mongoClient *mongo.Client, walletAddress string, NftCollections []NftCollection) error {
	dbCtx, close := context.WithTimeout(ctx, 10*time.Second)
	defer close()

	walletsCollection := mongoClient.Database("create_nft_tma").Collection("wallets")

	filter := bson.D{{Key: "_id", Value: walletAddress}}

	update := bson.D{{Key: "$set", Value: bson.D{{Key: "nft_collections", Value: NftCollections}}}}

	if _, updateErr := walletsCollection.UpdateOne(dbCtx, filter, update); updateErr != nil {
		return updateErr
	}

	return nil
}

func FindWalletInMongoByAddress(ctx context.Context, mongoClient *mongo.Client, address string) (*Wallet, error) {
	dbCtx, close := context.WithTimeout(ctx, 5*time.Second)
	defer close()

	walletsCollection := mongoClient.Database("create_nft_tma").Collection("wallets")

	var wallet *Wallet

	if findErr := walletsCollection.FindOne(dbCtx, bson.D{{Key: "_id", Value: address}}).Decode(&wallet); findErr != nil {
		return nil, findErr
	}

	return wallet, nil
}
