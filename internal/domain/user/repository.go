package user

import (
	"context"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, userID int64) (*User, error)
	CreateUser(ctx context.Context, user *User) error
}

//Основные коды ошибкок
