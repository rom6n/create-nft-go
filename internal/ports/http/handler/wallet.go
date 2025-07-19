package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/rom6n/create-nft-go/internal/database/nosql"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func GetWalletData(mongoClient *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		walletAddress := c.Query("wallet-address")

		if walletAddress == "" {
			return c.Status(fiber.StatusBadRequest).SendString("Wallet Address is required")
		}

		wallet, findErr := nosql.FindWalletInMongoByAddress(ctx, mongoClient, walletAddress)
		if findErr != nil {
			// добавляем кошелек в БД если его нет
			/*if findErr == mongo.ErrNoDocuments {
				// Получаем информацию о кошельке
				//
				// затем добавляем данные кошелька
				//
				// wallet := Wallet{ToDo}
				// nosql.AddWalletToMongo(ctx, mongoClient, wallet)
				// return ToDo
			}*/

			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("MongoDB find error: %v", findErr))
		}

		return c.Status(fiber.StatusFound).JSON(wallet)
	}
}

func UpdateWalletData(mongoClient *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		//ctx := c.Context()

		walletAddress := c.Query("wallet-address")
		if walletAddress == "" {
			return c.Status(fiber.StatusBadRequest).SendString("Wallet Address is required")
		}

		// получаем информацию о кошельке через TonApi
		//
		// Создаем на ее основе структуру Wallet
		// wallet := Wallet{ToDo}
		//
		// Обновляем БД
		//nosql.UpdateWalletInMongo(ctx, mongoClient, wallet)
		return c.Status(fiber.StatusForbidden).SendString("Функция / api еще не работает. ToDo")
	}
}
