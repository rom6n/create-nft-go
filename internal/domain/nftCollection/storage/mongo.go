package storage

import (
	nftcollection "github.com/rom6n/create-nft-go/internal/domain/nftCollection"
)

type nftCollectionRepo struct{}

func NewNftCollectionRepo() nftcollection.NftCollectionRepository {
	return &nftCollectionRepo{}
}
