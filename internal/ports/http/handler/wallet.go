package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/rom6n/create-nft-go/internal/database/nosql"
	"github.com/rom6n/create-nft-go/internal/ports/http/api"
	"github.com/tonkeeper/tonapi-go"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func GetWalletData(mongoClient *mongo.Client, tonApiClient *tonapi.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		walletAddress := c.Query("wallet-address")

		if walletAddress == "" {
			return c.Status(fiber.StatusBadRequest).SendString("Wallet Address is required")
		}

		wallet, findErr := nosql.FindWalletInMongoByAddress(ctx, mongoClient, walletAddress)
		if findErr != nil {
			// добавляем кошелек в БД если его нет
			if findErr == mongo.ErrNoDocuments {
				// Получаем информацию о кошельке
				nftItems, apiErr := api.GetWalletNftItems(ctx, tonApiClient, walletAddress)
				if apiErr != nil {
					return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Error while getting wallet nft items: %v", apiErr))
				}

				wallet := nosql.Wallet{
					Address:        walletAddress,
					NftItems:       *nftItems,
					NftCollections: nil,
				}

				nosql.AddWalletToMongo(ctx, mongoClient, &wallet)

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

		walletAddress := c.Query("wallet-address")
		if walletAddress == "" {
			return c.Status(fiber.StatusBadRequest).SendString("Wallet Address is required")
		}

		nftItems, apiErr := api.GetWalletNftItems(ctx, tonApiClient, walletAddress)
		if apiErr != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Error while getting wallet nft items: %v", apiErr))
		}

		if updateErr := nosql.UpdateWalletNftItemsInMongo(ctx, mongoClient, walletAddress, nftItems); updateErr != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Error update NFT items: %v", updateErr))
		}

		return c.Status(fiber.StatusOK).JSON(nftItems)
	}
}
