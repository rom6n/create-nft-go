package user

import (
	"context"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, userID int64) (*User, error)
}

//Основные коды ошибкок
