package walletservice

import (
	"context"

	"github.com/rom6n/create-nft-go/internal/domain/wallet"
	"github.com/rom6n/create-nft-go/internal/ports/http/api/ton"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type WalletServiceRepository interface {
	UpdateWalletNftItems(ctx context.Context, walletAddress string) (*[]wallet.NftItem, error)
	GetWalletByAddress(ctx context.Context, walletAddress string) (*wallet.Wallet, error)
}

type walletServiceRepo struct {
	TonApi     ton.TonApiRepository
	WalletRepo wallet.WalletRepository
}

func New(tonApi ton.TonApiRepository, walletRepo wallet.WalletRepository) WalletServiceRepository {
	return &walletServiceRepo{
		TonApi:     tonApi,
		WalletRepo: walletRepo,
	}
}

func (v *walletServiceRepo) createWallet(ctx context.Context, wallet *wallet.Wallet) error {
	return v.WalletRepo.AddWallet(ctx, wallet)
}

func (v *walletServiceRepo) UpdateWalletNftItems(ctx context.Context, walletAddress string) (*[]wallet.NftItem, error) {
	nftItems, apiErr := v.TonApi.GetWalletNftItems(ctx, walletAddress)
	if apiErr != nil {
		return nil, apiErr
	}

	if refreshErr := v.WalletRepo.UpdateWalletNftItems(ctx, walletAddress, nftItems); refreshErr != nil {
		return nil, refreshErr
	}

	return nftItems, nil
}

func (v *walletServiceRepo) GetWalletByAddress(ctx context.Context, walletAddress string) (*wallet.Wallet, error) {
	walletGotten, dbErr := v.WalletRepo.GetWalletByAddress(ctx, walletAddress)

	if dbErr != nil {
		// добавляем кошелек в БД если его нет
		if dbErr == mongo.ErrNoDocuments {
			// Получаем информацию о кошельке
			nftItems, apiErr := v.TonApi.GetWalletNftItems(ctx, walletAddress)
			if apiErr != nil {
				return nil, apiErr
			}

			wallet := wallet.Wallet{
				Address:        walletAddress,
				NftItems:       *nftItems,
				NftCollections: nil,
			}

			if createErr := v.createWallet(ctx, &wallet); createErr != nil {
				return nil, createErr
			}

			return &wallet, nil
		}

		return nil, dbErr
	}

	return walletGotten, nil
}
