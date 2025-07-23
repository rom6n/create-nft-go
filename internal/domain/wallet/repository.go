package wallet

import "context"

type WalletRepository interface {
	AddWallet(ctx context.Context, wallet *Wallet) error
	UpdateWalletNftItems(ctx context.Context, walletAddress string, nftItems *[]NftItem) error
	GetWalletByAddress(ctx context.Context, address string) (*Wallet, error)
}

//Основные коды ошибкок