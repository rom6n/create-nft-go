package handler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-faster/jx"
	"github.com/gofiber/fiber/v2"
	"github.com/rom6n/create-nft-go/internal/database/nosql"
	"github.com/tonkeeper/tonapi-go"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func GetWalletData(mongoClient *mongo.Client, tonApiClient *tonapi.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		apiCtx, close := context.WithTimeout(ctx, 30*time.Second)
		defer close()

		walletAddress := c.Query("wallet-address")

		if walletAddress == "" {
			return c.Status(fiber.StatusBadRequest).SendString("Wallet Address is required")
		}

		wallet, findErr := nosql.FindWalletInMongoByAddress(ctx, mongoClient, walletAddress)
		if findErr != nil {
			// добавляем кошелек в БД если его нет
			if findErr == mongo.ErrNoDocuments {
				// Получаем информацию о кошельке
				nfts, nftsErr := tonApiClient.GetAccountNftItems(apiCtx, tonapi.GetAccountNftItemsParams{
					AccountID:         walletAddress,
					IndirectOwnership: tonapi.OptBool{Value: true, Set: true},
				})

				if nftsErr != nil {
					return c.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("TonApi Get error: %v", nftsErr))
				}

				var nftItems []nosql.NftItem

				for _, items := range nfts.NftItems {
					meta := items.Metadata

					metadataName, nameErr := jx.DecodeBytes(meta["name"]).Str()
					metadataImage, imageErr := jx.DecodeBytes(meta["image"]).Str()
					metadataDescription, descriptionErr := jx.DecodeBytes(meta["description"]).Str()
					metadataExternalUrl, urlErr := jx.DecodeBytes(meta["external_url"]).Str()
					if nameErr != nil || imageErr != nil || descriptionErr != nil {
						log.Printf("‼️Error decoding NFTs metadata:\nName: %v\nImage: %v\nDesc: %v\nURL: %v\n", nameErr, imageErr, descriptionErr, urlErr)
					}

					clearMetadata := make(map[string]string)
					clearMetadata["name"] = metadataName
					clearMetadata["image"] = metadataImage
					clearMetadata["description"] = metadataDescription
					clearMetadata["external_url"] = metadataExternalUrl

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

				wallet := nosql.Wallet{
					Address:        walletAddress,
					NftItems:       nftItems,
					NftCollections: nil,
				}

				nosql.AddWalletToMongo(ctx, mongoClient, wallet)

				return c.Status(fiber.StatusFound).JSON(wallet)
			}

			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("MongoDB find error: %v", findErr))
		}

		return c.Status(fiber.StatusFound).JSON(wallet)
	}
}

func UpdateWalletNftItems(mongoClient *mongo.Client, tonApiClient *tonapi.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		apiCtx, close := context.WithTimeout(ctx, 30*time.Second)
		defer close()

		walletAddress := c.Query("wallet-address")
		if walletAddress == "" {
			return c.Status(fiber.StatusBadRequest).SendString("Wallet Address is required")
		}

		nfts, nftsErr := tonApiClient.GetAccountNftItems(apiCtx, tonapi.GetAccountNftItemsParams{
			AccountID:         walletAddress,
			IndirectOwnership: tonapi.OptBool{Value: true, Set: true},
		})

		if nftsErr != nil {
			return c.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("TonApi Get error: %v", nftsErr))
		}

		var nftItems []nosql.NftItem

		for _, items := range nfts.NftItems {
			meta := items.Metadata

			metadataName, nameErr := jx.DecodeBytes(meta["name"]).Str()
			metadataImage, imageErr := jx.DecodeBytes(meta["image"]).Str()
			metadataDescription, descriptionErr := jx.DecodeBytes(meta["description"]).Str()
			metadataExternalUrl, urlErr := jx.DecodeBytes(meta["external_url"]).Str()
			if nameErr != nil || imageErr != nil || descriptionErr != nil {
				log.Printf("‼️Error decoding NFTs metadata:\nName: %v\nImage: %v\nDesc: %v\nURL: %v\n", nameErr, imageErr, descriptionErr, urlErr)
			}

			clearMetadata := make(map[string]string)
			clearMetadata["name"] = metadataName
			clearMetadata["image"] = metadataImage
			clearMetadata["description"] = metadataDescription
			clearMetadata["external_url"] = metadataExternalUrl

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

		if updateErr := nosql.UpdateWalletNftItemsInMongo(ctx, mongoClient, walletAddress, nftItems); updateErr != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Error update NFT items: %v", updateErr))
		}

		return c.Status(fiber.StatusOK).JSON(nftItems)
	}
}
