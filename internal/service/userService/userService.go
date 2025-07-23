package userservice

import (
	"context"

	"github.com/rom6n/create-nft-go/internal/domain/user"
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
	return v.UserRepo.GetUserByID(ctx, userID)
}
