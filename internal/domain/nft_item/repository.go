package nftitem

import (
	"context"

	"github.com/google/uuid"
)

type NftItemRepository interface {
	CreateNftItem(ctx context.Context, nftItem *NftItem) error
	GetNftItemsByOwnerUuid(ctx context.Context, uuid uuid.UUID) ([]NftItem, error)
	GetNftItemByAddress(ctx context.Context, nftItemAddress string) (*NftItem, error)
	DeleteNftItem(ctx context.Context, nftItemAddress string) error
}

//Основные коды ошибкок
