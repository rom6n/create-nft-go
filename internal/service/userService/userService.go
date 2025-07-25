package userservice

import (
	"context"

	"github.com/google/uuid"
	"github.com/rom6n/create-nft-go/internal/domain/user"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type UserServiceRepository interface {
	GetUserByID(ctx context.Context, userID int64) (*user.User, error)
}

type userServiceRepo struct {
	UserRepo user.UserRepository
}

func New(userRepo user.UserRepository) UserServiceRepository {
	return &userServiceRepo{UserRepo: userRepo}
}

func (v *userServiceRepo) GetUserByID(ctx context.Context, userID int64) (*user.User, error) {
	foundUser, dbErr := v.UserRepo.GetUserByID(ctx, userID)
	if dbErr != nil {
		if dbErr == mongo.ErrNoDocuments {
			newUuid := uuid.New()
			user := user.NewUser(newUuid, userID, 1, "user")
			createErr := v.UserRepo.CreateUser(ctx, &user)
			return &user, createErr
		}
		return nil, dbErr
	}

	return foundUser, nil
}
