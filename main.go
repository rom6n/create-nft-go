package main

import (
	"context"
	"log"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	nftcollectionrepo "github.com/rom6n/create-nft-go/internal/domain/nft_collection/storage"
	userRepo "github.com/rom6n/create-nft-go/internal/domain/user/storage"
	walletRepo "github.com/rom6n/create-nft-go/internal/domain/wallet/storage"
	"github.com/rom6n/create-nft-go/internal/ports/http/api/ton"
	"github.com/rom6n/create-nft-go/internal/ports/http/handler"
	deploynftcollection "github.com/rom6n/create-nft-go/internal/service/deploy_nft_collection"
	nftcollectionservice "github.com/rom6n/create-nft-go/internal/service/nft_collection_service"
	userservice "github.com/rom6n/create-nft-go/internal/service/user_service"
	walletservice "github.com/rom6n/create-nft-go/internal/service/wallet_service"
	"github.com/rom6n/create-nft-go/internal/storage"
	marketutils "github.com/rom6n/create-nft-go/internal/utils/contract_utils/market_utils"
	nftcollectionutils "github.com/rom6n/create-nft-go/internal/utils/contract_utils/nft_collection_utils"
	nftitemutils "github.com/rom6n/create-nft-go/internal/utils/contract_utils/nft_item_utils"
	"github.com/rom6n/create-nft-go/internal/utils/tonutil"
)

func main() {
	ctx := context.Background()
	if loadErr := godotenv.Load(); loadErr != nil {
		log.Fatal("‼️ Error loading .env file")
	}

	// ---------------------------------- Init -----------------------------------------

	privateKey := tonutil.GetTestPrivateKey()
	liteClient, liteclientApi := tonutil.GetLiteClient(ctx)
	marketplaceContractAddress := marketutils.GetMarketplaceContractAddress()
	nftCollectionContractCode := nftcollectionutils.GetNftCollectionContractCode()
	nftItemContractCode := nftitemutils.GetNftItemContractCode()
	marketplaceContractCode := marketutils.GetMarketplaceContractCode()
	tonapiClient := ton.NewTonapiClient()

	databaseClient := storage.NewMongoClient()
	defer databaseClient.Disconnect(ctx)

	// ---------------------------------- Repo -------------------------------------------

	walletRepo := walletRepo.NewWalletRepo(databaseClient, walletRepo.WalletRepoCfg{
		DBName:         "create-nft-tma",
		CollectionName: "wallets",
		Timeout:        15 * time.Second,
	})

	nftCollectionRepo := nftcollectionrepo.NewNftCollectionRepo(databaseClient, nftcollectionrepo.NftCollectionRepoCfg{
		DBName:         "create-nft-tma",
		CollectionName: "nft-collections",
		Timeout:        15 * time.Second,
	})

	userRepo := userRepo.NewUserRepo(databaseClient, userRepo.UserRepoCfg{
		DBName:         "create-nft-tma",
		CollectionName: "users",
		Timeout:        15 * time.Second,
	})

	userServiceRepo := userservice.New(userRepo)

	deployNftCollectionServiceRepo := deploynftcollection.New(deploynftcollection.DeployNftCollectionServiceCfg{
		NftCollectionRepo:          nftCollectionRepo,
		UserRepo:                   userRepo,
		PrivateKey:                 privateKey,
		LiteClient:                 liteClient,
		LiteclientApi:              liteclientApi,
		MarketplaceContractAddress: marketplaceContractAddress,
		NftCollectionContractCode:  nftCollectionContractCode,
		NftItemContractCode:        nftItemContractCode,
	})

	nftCollectionServiceRepo := nftcollectionservice.New(nftcollectionservice.NftCollectionServiceCfg{
		NftCollectionRepo:       nftCollectionRepo,
		UserRepo:                userRepo,
		PrivateKey:              privateKey,
		LiteClient:              liteClient,
		LiteclientApi:           liteclientApi,
		MarketplaceContractCode: marketplaceContractCode,
	})

	tonApiRepo := ton.NewTonApiRepo(tonapiClient, 30*time.Second)

	walletServiceRepo := walletservice.New(tonApiRepo, walletRepo)

	// -------------------------------- Handlers -----------------------------------------------

	WalletHandler := handler.WalletHandler{
		WalletServiceRepo: walletServiceRepo,
	}

	UserHandler := handler.UserHandler{
		UserService: userServiceRepo,
	}

	NftCollectionHandler := handler.NftCollectionHandler{
		NftCollectionService:       nftCollectionServiceRepo,
		DeployNftCollectionService: deployNftCollectionServiceRepo,
	}

	// ------------------------------- App & Routes --------------------------------------

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
