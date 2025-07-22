package user

import (
	"context"
	"time"
	"github.com/rom6n/create-nft-go/internal/domain/user"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type mongoUserRepository struct {
	client         *mongo.Client
	dbName         string
	collectionName string
	timeout        time.Duration
}

type UserRepositoryCfg struct {
	DBName         string
	CollectionName string
	Timeout        time.Duration
}

func NewUserRepository(client *mongo.Client, cfg UserRepositoryCfg) user.UserRepository {
	return NewMongoUserRepository(
		client,
		cfg.DBName,
		cfg.CollectionName,
		cfg.Timeout,
	)
}

func (r *mongoUserRepository) GetContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, r.timeout)
}

func (r *mongoUserRepository) GetCollection() *mongo.Collection {
	return r.client.Database(r.dbName).Collection(r.collectionName)
}

func NewMongoUserRepository(client *mongo.Client, dbName string, collectionName string, timeout time.Duration) user.UserRepository {
	return &mongoUserRepository{
		client:         client,
		dbName:         dbName,
		collectionName: collectionName,
		timeout:        timeout,
	}
}

func (r *mongoUserRepository) GetUserByID(ctx context.Context, userID int64) (*user.User, error) {
	dbCtx, cancel := r.GetContext(ctx)
	defer cancel()

	userCollection := r.GetCollection()

	var user user.User

	if findErr := userCollection.FindOne(dbCtx, bson.D{{Key: "_id", Value: userID}}).Decode(&user); findErr != nil {
		return nil, findErr
	}

	return &user, nil
}
