package nftcollection

import (
	"context"
)

type NftCollectionRepository interface {
	CreateNftCollection(ctx context.Context, collection *NftCollection) error
	DeleteNftCollection(ctx context.Context, collectionAddress string) error
	GetNftCollectionByAddress(ctx context.Context, collectionAddress string) (*NftCollection, error)
}

