package user

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rom6n/create-nft-go/internal/domain/user"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type mongoUserRepo struct {
	client         *mongo.Client
	dbName         string
	collectionName string
	timeout        time.Duration
}

type UserRepoCfg struct {
	DBName         string
	CollectionName string
	Timeout        time.Duration
}

func NewUserRepo(client *mongo.Client, cfg UserRepoCfg) user.UserRepository {
	return &mongoUserRepo{
		client:         client,
		dbName:         cfg.DBName,
		collectionName: cfg.CollectionName,
		timeout:        cfg.Timeout,
	}
}

func (r *mongoUserRepo) getContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, r.timeout)
}

func (r *mongoUserRepo) getCollection() *mongo.Collection {
	return r.client.Database(r.dbName).Collection(r.collectionName)
}

func (r *mongoUserRepo) GetUserByID(ctx context.Context, userID int64) (*user.User, error) {
	dbCtx, cancel := r.getContext(ctx)
	defer cancel()

	userCollection := r.getCollection()

	var user user.User

	if findErr := userCollection.FindOne(dbCtx, bson.D{{Key: "id", Value: userID}}).Decode(&user); findErr != nil {
		return nil, findErr
	}

	return &user, nil
}

func (v *mongoUserRepo) CreateUser(ctx context.Context, user *user.User) error {
	dbCtx, cancel := v.getContext(ctx)
	defer cancel()

	collection := v.getCollection()

	_, insertErr := collection.InsertOne(dbCtx, *user)
	return insertErr
}

func (v *mongoUserRepo) UpdateUserBalance(ctx context.Context, userUuid uuid.UUID, newNanoTon uint64) error {
	dbCtx, cancel := v.getContext(ctx)
	defer cancel()

	collection := v.getCollection()
	_, updErr := collection.UpdateOne(dbCtx, bson.D{{Key: "_id", Value: userUuid}}, bson.D{{Key: "$set", Value: bson.D{{Key: "nano_ton", Value: newNanoTon}}}})
	if updErr != nil {
		return fmt.Errorf("error update user uuid %v's balance: %v", userUuid, updErr)
	}
	return nil
}
