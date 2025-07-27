package main

import (
	"context"
	"log"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	nftcollectionrepo "github.com/rom6n/create-nft-go/internal/domain/nftCollection/storage"
	userRepo "github.com/rom6n/create-nft-go/internal/domain/user/storage"
	walletRepo "github.com/rom6n/create-nft-go/internal/domain/wallet/storage"
	"github.com/rom6n/create-nft-go/internal/ports/http/api/ton"
	"github.com/rom6n/create-nft-go/internal/ports/http/handler"
	nftcollectionservice "github.com/rom6n/create-nft-go/internal/service/nftCollectionService"
	userservice "github.com/rom6n/create-nft-go/internal/service/userService"
	walletservice "github.com/rom6n/create-nft-go/internal/service/walletService"
	"github.com/rom6n/create-nft-go/internal/storage"
	"github.com/rom6n/create-nft-go/internal/util/tonutil"
)

func main() {
	ctx := context.Background()
	if loadErr := godotenv.Load(); loadErr != nil {
		log.Fatal("‼️ Error loading .env file")
	}

	privateKey := tonutil.GetTestPrivateKey()

	databaseClient := storage.NewMongoClient()
	defer databaseClient.Disconnect(ctx)

	walletRepo := walletRepo.NewWalletRepo(databaseClient, walletRepo.WalletRepoCfg{
		DBName:         "create-nft-tma",
		CollectionName: "wallets",
		Timeout:        10 * time.Second,
	})

	nftCollectionRepo := nftcollectionrepo.NewNftCollectionRepo()

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

	NftCollectionHandler := handler.NftCollectionHandler{
		NftCollectionService: nftcollectionservice.New(nftCollectionRepo, userRepo, privateKey),
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
	nftCollectionApi := api.Group("/collection")

	walletApi.Get("/get-wallet-data", WalletHandler.GetWalletData())

	walletApi.Get("/refresh-wallet-nft-items", WalletHandler.RefreshWalletNftItems()) // В будущем поменять на PUT

	userApi.Get("/get-user-data", UserHandler.GetUserData())

	nftCollectionApi.Get("/deploy-market", NftCollectionHandler.DeployMarketContract())

	nftCollectionApi.Get("/deploy-collection-test", NftCollectionHandler.DeployNftCollection())

	app.Listen(":2000")
}
