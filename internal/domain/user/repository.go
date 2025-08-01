package user

import (
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, userID int64) (*User, error)
	CreateUser(ctx context.Context, user *User) error
	UpdateUserBalance(ctx context.Context, userUuid uuid.UUID, newNanoTon uint64) error
}

//Основные коды ошибкок
