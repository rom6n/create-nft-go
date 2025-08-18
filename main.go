package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	nftcollectionrepo "github.com/rom6n/create-nft-go/internal/domain/nft_collection/storage"
	nftitemRepo "github.com/rom6n/create-nft-go/internal/domain/nft_item/storage"
	userRepo "github.com/rom6n/create-nft-go/internal/domain/user/storage"
	walletRepo "github.com/rom6n/create-nft-go/internal/domain/wallet/storage"
	"github.com/rom6n/create-nft-go/internal/ports/http/api/ton"
	"github.com/rom6n/create-nft-go/internal/ports/http/handler"
	deploynftcollection "github.com/rom6n/create-nft-go/internal/service/deploy_nft_collection"
	marketplacecontractservice "github.com/rom6n/create-nft-go/internal/service/marketplace_contract_service"
	mintnftitem "github.com/rom6n/create-nft-go/internal/service/mint_nft_item"
	nftcollectionservice "github.com/rom6n/create-nft-go/internal/service/nft_collection_service"
	userservice "github.com/rom6n/create-nft-go/internal/service/user_service"
	walletservice "github.com/rom6n/create-nft-go/internal/service/wallet_service"
	withdrawnftcollection "github.com/rom6n/create-nft-go/internal/service/withdraw_nft_collection"
	withdrawnftitem "github.com/rom6n/create-nft-go/internal/service/withdraw_nft_item"
	"github.com/rom6n/create-nft-go/internal/storage"
	marketutils "github.com/rom6n/create-nft-go/internal/utils/contract_utils/market_utils"
	nftcollectionutils "github.com/rom6n/create-nft-go/internal/utils/contract_utils/nft_collection_utils"
	nftitemutils "github.com/rom6n/create-nft-go/internal/utils/contract_utils/nft_item_utils"
	"github.com/rom6n/create-nft-go/internal/utils/tonutil"
)

func main() {
	ctx := context.Background()

	if loadErr := godotenv.Load(); loadErr != nil {
		log.Println("Using system environment variables")
	}

	// ---------------------------------- Init -----------------------------------------

	privateKey := tonutil.GetPrivateKey()
	testnetLiteClient, testnetLiteApi := tonutil.GetTestnetLiteClient(ctx)
	mainnetLiteClient, mainnetLiteApi := tonutil.GetMainnetLiteClient(ctx)
	testnetMarketplaceContractAddress := marketutils.GetTestnetMarketplaceContractAddress()
	mainnetMarketplaceContractAddress := marketutils.GetMainnetMarketplaceContractAddress()
	nftCollectionContractCode := nftcollectionutils.GetNftCollectionContractCode()
	nftItemContractCode := nftitemutils.GetNftItemContractCode()
	marketplaceContractCode := marketutils.GetMarketplaceContractCode()
	tonapiClient := ton.NewTonapiClient()
	testnetWallet := tonutil.GetTestnetWallet(testnetLiteApi)
	mainnetWallet := tonutil.GetMainnetWallet(mainnetLiteApi)

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

	nftItemRepo := nftitemRepo.NewNftItemRepo(databaseClient, nftitemRepo.NftItemRepoCfg{
		DBName:         "create-nft-tma",
		CollectionName: "nft-items",
		Timeout:        15 * time.Second,
	})

	userServiceRepo := userservice.New(userservice.UserServiceCfg{
		UserRepo:          userRepo,
		NftCollectionRepo: nftCollectionRepo,
		NftItemRepo:       nftItemRepo,
		Timeout:           30 * time.Second,
	})

	deployNftCollectionServiceRepo := deploynftcollection.New(deploynftcollection.DeployNftCollectionServiceCfg{
		NftCollectionRepo:                 nftCollectionRepo,
		UserRepo:                          userRepo,
		PrivateKey:                        privateKey,
		TestnetLiteClient:                 testnetLiteClient,
		MainnetLiteClient:                 mainnetLiteClient,
		TestnetLiteApi:                    testnetLiteApi,
		MainnetLiteApi:                    mainnetLiteApi,
		TestnetMarketplaceContractAddress: testnetMarketplaceContractAddress,
		MainnetMarketplaceContractAddress: mainnetMarketplaceContractAddress,
		NftCollectionContractCode:         nftCollectionContractCode,
		NftItemContractCode:               nftItemContractCode,
		Timeout:                           30 * time.Second,
	})

	nftCollectionServiceRepo := nftcollectionservice.New(nftcollectionservice.NftCollectionServiceCfg{
		NftCollectionRepo:       nftCollectionRepo,
		UserRepo:                userRepo,
		PrivateKey:              privateKey,
		LiteClient:              testnetLiteClient,
		LiteclientApi:           testnetLiteApi,
		MarketplaceContractCode: marketplaceContractCode,
	})

	mintNftItemServiceRepo := mintnftitem.New(mintnftitem.MintNftItemServiceCfg{
		NftCollectionRepo:                 nftCollectionRepo,
		NftItemRepo:                       nftItemRepo,
		UserRepo:                          userRepo,
		NftItemCode:                       nftItemContractCode,
		TestnetLiteClient:                 testnetLiteClient,
		MainnetLiteClient:                 mainnetLiteClient,
		TestnetLiteApi:                    testnetLiteApi,
		MainnetLiteApi:                    mainnetLiteApi,
		TestnetMarketplaceContractAddress: testnetMarketplaceContractAddress,
		MainnetMarketplaceContractAddress: mainnetMarketplaceContractAddress,
		PrivateKey:                        privateKey,
		Timeout:                           30 * time.Second,
	})

	marketplaceContractServiceRepo := marketplacecontractservice.New(marketplacecontractservice.MarketplaceContractServiceCfg{
		TestnetLiteClient:                 testnetLiteClient,
		MainnetLiteClient:                 mainnetLiteClient,
		TestnetMarketplaceContractAddress: testnetMarketplaceContractAddress,
		MainnetMarketplaceContractAddress: mainnetMarketplaceContractAddress,
		TestnetLiteApi:                    testnetLiteApi,
		MainnetLiteApi:                    mainnetLiteApi,
		TestnetWallet:                     testnetWallet,
		MainnetWallet:                     mainnetWallet,
		PrivateKey:                        privateKey,
		MarketplaceContractCode:           marketplaceContractCode,
		Timeout:                           30 * time.Second,
	})

	withdrawNftCollectionServiceRepo := withdrawnftcollection.New(withdrawnftcollection.WithdrawNftCollectionServiceCfg{
		NftCollectionRepo:                 nftCollectionRepo,
		UserRepo:                          userRepo,
		PrivateKey:                        privateKey,
		TestnetLiteClient:                 testnetLiteClient,
		MainnetLiteClient:                 mainnetLiteClient,
		TestnetLiteApi:                    testnetLiteApi,
		MainnetLiteApi:                    mainnetLiteApi,
		TestnetMarketplaceContractAddress: testnetMarketplaceContractAddress,
		MainnetMarketplaceContractAddress: mainnetMarketplaceContractAddress,
		Timeout:                           30 * time.Second,
	})

	withdrawNftItemServiceRepo := withdrawnftitem.New(withdrawnftitem.WithdrawNftItemServiceCfg{
		NftItemRepo:                       nftItemRepo,
		UserRepo:                          userRepo,
		PrivateKey:                        privateKey,
		TestnetLiteClient:                 testnetLiteClient,
		MainnetLiteClient:                 mainnetLiteClient,
		TestnetLiteApi:                    testnetLiteApi,
		MainnetLiteApi:                    mainnetLiteApi,
		TestnetMarketplaceContractAddress: testnetMarketplaceContractAddress,
		MainnetMarketplaceContractAddress: mainnetMarketplaceContractAddress,
		Timeout:                           30 * time.Second,
	})

	tonApiRepo := ton.NewTonApiRepo(tonapiClient, 30*time.Second)

	walletServiceRepo := walletservice.New(tonApiRepo, walletRepo)

	// -------------------------------- Handlers -----------------------------------------------

	walletHandler := handler.WalletHandler{
		WalletServiceRepo: walletServiceRepo,
	}

	userHandler := handler.UserHandler{
		UserService: userServiceRepo,
	}

	nftCollectionHandler := handler.NftCollectionHandler{
		NftCollectionService:         nftCollectionServiceRepo,
		DeployNftCollectionService:   deployNftCollectionServiceRepo,
		WithdrawNftCollectionService: withdrawNftCollectionServiceRepo,
	}

	nftItemHandler := handler.NftItemHandler{
		MintNftItemService:     mintNftItemServiceRepo,
		WithdrawNftItemService: withdrawNftItemServiceRepo,
	}

	marketplaceHandler := handler.MarketplaceContractHandler{
		MarketplaceContractService: marketplaceContractServiceRepo,
	}

	// ------------------------------- App & Routes --------------------------------------

	app := fiber.New(fiber.Config{
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})

	app.Use(logger.New())

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("pong")
	})

	app.Get("/favicon.ico", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	api := app.Group("/api")
	walletApi := api.Group("/wallet")
	userApi := api.Group("/user")
	nftCollectionApi := api.Group("/nft-collection")
	nftItemApi := api.Group("/nft-item")
	marketApi := api.Group("/market")

	walletApi.Get("/get-wallet-data", walletHandler.GetWalletData())
	walletApi.Get("/refresh-wallet-nft-items", walletHandler.RefreshWalletNftItems()) // В будущем поменять на PUT

	marketApi.Get("/deploy", marketplaceHandler.DeployMarketContract())            // В будущем поменять на POST
	marketApi.Get("/deposit", marketplaceHandler.DepositMarket())                  // В будущем поменять на POST
	marketApi.Get("/withdraw", marketplaceHandler.WithdrawTonFromMarketContract()) // В будущем поменять на POST

	userApi.Get("/:id", userHandler.GetUserData())
	userApi.Get("/nft-collections/:id", userHandler.GetUserNftCollections())
	userApi.Get("/nft-items/:id", userHandler.GetUserNftItems())

	nftCollectionApi.Get("/deploy", nftCollectionHandler.DeployNftCollection())              // В будущем поменять на POST
	nftCollectionApi.Get("/withdraw/:address", nftCollectionHandler.WithdrawNftCollection()) // В будущем поменять на POST

	nftItemApi.Get("/mint", nftItemHandler.MintNftItem())                  // В будущем поменять на POST
	nftItemApi.Get("/withdraw/:address", nftItemHandler.WithdrawNftItem()) // В будущем поменять на POST

	go func() {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080" // дефолт для локального запуска
		}
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// ------------------------- Graceful shutdown --------------------------------

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt)

	<-stop
	shutdownTimeSecond := 3 * time.Second // basically 35 seconds
	shutdownTime := 4                     // basically 40 seconds

	ctxShutdown, cancel := context.WithTimeout(ctx, shutdownTimeSecond)
	defer cancel()

	if shotdownErr := app.ShutdownWithContext(ctxShutdown); shotdownErr != nil {
		log.Fatalf("Error shutting down server: %v. Forced shutdown", shotdownErr)
	}

	for i := shutdownTime; i > 0; i -= 1 {
		log.Printf("🕒 Shutting down in %v seconds...\n", i)
		time.Sleep(1 * time.Second)
	}

	log.Println("✅ Server shutdown successfully")
}
