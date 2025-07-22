package main

import (
	"context"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	userRepo "github.com/rom6n/create-nft-go/internal/domain/user/storage"
	walletRepo "github.com/rom6n/create-nft-go/internal/domain/wallet/storage"
	"github.com/rom6n/create-nft-go/internal/ports/http/api/ton"
	"github.com/rom6n/create-nft-go/internal/ports/http/handler"
	"github.com/rom6n/create-nft-go/internal/storage"
)

func main() {
	ctx := context.Background()

	DatabaseClient := storage.NewMongoClient()
	defer DatabaseClient.Disconnect(ctx)

	WalletRepository := walletRepo.NewWalletRepository(DatabaseClient, walletRepo.WalletRepositoryCfg{
		DBName:           "create-nft-tma",
		WalletCollection: "wallets",
		Timeout:          10 * time.Second,
	})

	UserRepository := userRepo.NewUserRepository(DatabaseClient, userRepo.UserRepositoryCfg{
		DBName:         "create-nft-tma",
		CollectionName: "users",
		Timeout:        10 * time.Second,
	})

	tonapiClient := ton.NewTonapiClient()
	TonApiRepository := ton.NewTonApiRepository(tonapiClient, 30*time.Second)
	WalletHandler := handler.WalletHandler{
		WalletDB: WalletRepository,
		TonApi:   TonApiRepository,
	}

	UserHandler := handler.UserHandler{
		UserDB: UserRepository,
	}

	app := fiber.New(fiber.Config{
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})

	app.Use(logger.New())

	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("pong")
	})

	app.Get("/favicon.ico", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	api := app.Group("/api")
	walletApi := api.Group("/wallet")
	userApi := api.Group("/user")

	walletApi.Get("/get-wallet-data", WalletHandler.GetWalletData())

	walletApi.Get("/refresh-wallet-nft-items", WalletHandler.RefreshWalletNftItems()) // В будущем поменять на PUT

	userApi.Get("/get-user-data", UserHandler.GetUserData())

	app.Listen(":2000")
}
