package nftcollection

import (
	"context"

	"github.com/google/uuid"
)

type NftCollectionRepository interface {
	GetCollectionByOwnerUuid(ctx context.Context, ownerUuid uuid.UUID)
}
