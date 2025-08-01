package nftcollection

import (
	"context"

	"github.com/google/uuid"
)

type NftCollectionRepository interface {
	CreateCollection(ctx context.Context, collection *NftCollection, ownerUuid uuid.UUID) error
	DeleteCollection(ctx context.Context, collectionAddress string) error
}

//GetCollectionByOwnerUuid(ctx context.Context, ownerUuid uuid.UUID)
