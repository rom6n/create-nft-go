package nftitem

import (
	"context"

	"github.com/google/uuid"
)

type NftItemRepository interface {
	//GetNftItemByAddress(ctx context.Context, address string) (*NftItem, error)
	CreateNftItem(ctx context.Context, nftItem *NftItem) error
	GetNftItemsByOwnerUuid(ctx context.Context, uuid uuid.UUID) ([]NftItem, error)
}

//Основные коды ошибкок
