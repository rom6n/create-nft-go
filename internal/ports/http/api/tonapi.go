package api

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/rom6n/create-nft-go/internal/database/nosql"
	"github.com/rom6n/create-nft-go/internal/encoding/jsonx"
	"github.com/tonkeeper/tonapi-go"
)

func NewTonApiClient() *tonapi.Client {
	if loadErr := godotenv.Load(); loadErr != nil {
		log.Fatal("Error loading .env file")
	}

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

func GetWalletNftItems(ctx context.Context, tonApiClient *tonapi.Client, walletAddress string) (*[]nosql.NftItem, error) {
	apiCtx, close := context.WithTimeout(ctx, 30*time.Second)
	defer close()

	nfts, nftsErr := tonApiClient.GetAccountNftItems(apiCtx, tonapi.GetAccountNftItemsParams{
		AccountID:         walletAddress,
		IndirectOwnership: tonapi.OptBool{Value: true, Set: true},
	})
	if nftsErr != nil {
		return nil, nftsErr
	}

	var nftItems []nosql.NftItem

	for _, items := range nfts.NftItems {
		clearMetadata, decodeErr := jsonx.DecodeAndPackNftItemMetadata(items.Metadata)
		if decodeErr != nil {
			return nil, decodeErr
		}

		collectionName := items.Collection.Value.Name

		nftItems = append(nftItems, nosql.NftItem{
			Address:           items.Address,
			Index:             items.Index,
			CollectionAddress: items.Collection.Value.Address,
			CollectionName:    collectionName,
			Owner:             items.Owner.Value.Address,
			Metadata:          clearMetadata,
		})
	}

	return &nftItems, nil
}
