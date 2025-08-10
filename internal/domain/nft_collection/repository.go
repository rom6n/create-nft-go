package nftcollection

import (
	"context"

	"github.com/google/uuid"
)

type NftCollectionRepository interface {
	CreateNftCollection(ctx context.Context, collection *NftCollection) error
	DeleteNftCollection(ctx context.Context, collectionAddress string) error
	GetNftCollectionByAddress(ctx context.Context, collectionAddress string) (*NftCollection, error)
	GetNftCollectionsByOwnerUuid(ctx context.Context, uuid uuid.UUID) ([]NftCollection, error)
}

