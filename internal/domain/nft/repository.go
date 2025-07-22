package nft

import "context"

type NftItemRepository interface {
	GetNftItemByAddress(ctx context.Context, address string) (*NftItem, error)
}

//Основные коды ошибкок
