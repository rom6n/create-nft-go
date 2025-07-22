package ton

import (
	"context"
	"github.com/rom6n/create-nft-go/internal/domain/wallet"
)

type TonApiRepository interface {
	GetWalletNftItems(ctx context.Context, walletAddress string) (*[]wallet.NftItem, error)
}