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
	userservice "github.com/rom6n/create-nft-go/internal/service/userService"
	walletservice "github.com/rom6n/create-nft-go/internal/service/walletService"
	"github.com/rom6n/create-nft-go/internal/storage"
)

func main() {
	ctx := context.Background()

	databaseClient := storage.NewMongoClient()
	defer databaseClient.Disconnect(ctx)

	walletRepo := walletRepo.NewWalletRepo(databaseClient, walletRepo.WalletRepoCfg{
		DBName:         "create-nft-tma",
		CollectionName: "wallets",
		Timeout:        10 * time.Second,
	})

	userRepo := userRepo.NewUserRepo(databaseClient, userRepo.UserRepoCfg{
		DBName:         "create-nft-tma",
		CollectionName: "users",
		Timeout:        10 * time.Second,
	})

	userServiceRepo := userservice.New(userRepo)

	tonapiClient := ton.NewTonapiClient()
	tonApiRepository := ton.NewTonApiRepository(tonapiClient, 30*time.Second)

	walletServiceRepo := walletservice.New(tonApiRepository, walletRepo)

	WalletHandler := handler.WalletHandler{
		WalletServiceRepo: walletServiceRepo,
	}

	UserHandler := handler.UserHandler{
		UserService: userServiceRepo,
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
