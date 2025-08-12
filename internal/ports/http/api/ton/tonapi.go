package ton

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/rom6n/create-nft-go/internal/domain/wallet"
	"github.com/rom6n/create-nft-go/internal/utils/jsonx"
	"github.com/tonkeeper/tonapi-go"
)

type TonApiRepository interface {
	GetWalletNftItems(ctx context.Context, walletAddress string) ([]wallet.NftItem, error)
}

type TonapiTonApiCfg struct {
	Timeout time.Duration
}

type tonapiTonApiRepo struct {
	client  *tonapi.Client
	timeout time.Duration
}

func NewTonapiClient() *tonapi.Client {
	token := os.Getenv("TONAPI_TOKEN")
	if token == "" {
		log.Fatal("Error. Add TonApi token to env.")
	}

	client, err := tonapi.NewClient(tonapi.TonApiURL, tonapi.WithToken(token))
	if err != nil {
		log.Fatalf("TonApi connection error: %v\n", err)
	}

	return client
}

func NewTonApiRepo(tonapiClient *tonapi.Client, timeout time.Duration) TonApiRepository {
	return &tonapiTonApiRepo{
		client:  tonapiClient,
		timeout: timeout,
	}
}

func (r *tonapiTonApiRepo) getContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, r.timeout)
}

func (r *tonapiTonApiRepo) GetWalletNftItems(ctx context.Context, walletAddress string) ([]wallet.NftItem, error) {
	apiCtx, cancel := r.getContext(ctx)
	defer cancel()

	nfts, nftsErr := r.client.GetAccountNftItems(apiCtx, tonapi.GetAccountNftItemsParams{
		AccountID:         walletAddress,
		IndirectOwnership: tonapi.OptBool{Value: true, Set: true},
	})
	if nftsErr != nil {
		return nil, nftsErr
	}

	nftItems := make([]wallet.NftItem, 0, len(nfts.NftItems))

	for _, item := range nfts.NftItems {
		nftMetadata, decodeErr := jsonx.DecodeAndPackNftItemMetadata(item.Metadata)
		if decodeErr != nil {
			return nil, decodeErr
		}

		nftItem := wallet.NftItem{
			Address:  item.Address,
			Index:    item.Index,
			Metadata: nftMetadata,
		}

		if item.Collection.Set {
			nftItem.CollectionAddress = item.Collection.Value.Address
			nftItem.CollectionName = item.Collection.Value.Name
		}

		if item.Owner.Set {
			nftItem.Owner = item.Owner.Value.Address
		}

		nftItems = append(nftItems, nftItem)
	}

	return nftItems, nil
}
