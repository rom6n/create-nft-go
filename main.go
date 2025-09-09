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
	"github.com/rom6n/create-nft-go/internal/service/withdraw_user_ton"
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
	//testnetTonapiClient := ton.NewTestnetTonapiClient()
	testnetWallet := tonutil.GetTestnetWallet(testnetLiteApi)
	mainnetWallet := tonutil.GetMainnetWallet(mainnetLiteApi)
	//streamingApi := tonutil.GetStreamingApi()
	//testnetStreamingApi := tonutil.GetTestnetStreamingApi()

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
		NftCollectionRepo:         nftCollectionRepo,
		UserRepo:                  userRepo,
		PrivateKey:                privateKey,
		TestnetLiteClient:         testnetLiteClient,
		MainnetLiteClient:         mainnetLiteClient,
		TestnetLiteApi:            testnetLiteApi,
		MainnetLiteApi:            mainnetLiteApi,
		TestnetWallet:             testnetWallet,
		MainnetWallet:             mainnetWallet,
		NftCollectionContractCode: nftCollectionContractCode,
		NftItemContractCode:       nftItemContractCode,
		Timeout:                   30 * time.Second,
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
		NftCollectionRepo: nftCollectionRepo,
		NftItemRepo:       nftItemRepo,
		UserRepo:          userRepo,
		NftItemCode:       nftItemContractCode,
		TestnetLiteClient: testnetLiteClient,
		MainnetLiteClient: mainnetLiteClient,
		TestnetLiteApi:    testnetLiteApi,
		MainnetLiteApi:    mainnetLiteApi,
		TestnetWallet:     testnetWallet,
		MainnetWallet:     mainnetWallet,
		PrivateKey:        privateKey,
		Timeout:           30 * time.Second,
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
		NftCollectionRepo: nftCollectionRepo,
		UserRepo:          userRepo,
		PrivateKey:        privateKey,
		TestnetLiteClient: testnetLiteClient,
		MainnetLiteClient: mainnetLiteClient,
		TestnetLiteApi:    testnetLiteApi,
		MainnetLiteApi:    mainnetLiteApi,
		TestnetWallet:     testnetWallet,
		MainnetWallet:     mainnetWallet,
		Timeout:           30 * time.Second,
	})

	withdrawNftItemServiceRepo := withdrawnftitem.New(withdrawnftitem.WithdrawNftItemServiceCfg{
		NftItemRepo:       nftItemRepo,
		UserRepo:          userRepo,
		PrivateKey:        privateKey,
		TestnetLiteClient: testnetLiteClient,
		MainnetLiteClient: mainnetLiteClient,
		TestnetLiteApi:    testnetLiteApi,
		MainnetLiteApi:    mainnetLiteApi,
		TestnetWallet:     testnetWallet,
		MainnetWallet:     mainnetWallet,
		Timeout:           30 * time.Second,
	})

	withdrawUserRepo := withdraw_user_ton.New(withdraw_user_ton.WithdrawUserTonCfg{
		UserRepo:          userRepo,
		TestnetLiteClient: testnetLiteClient,
		MainnetLiteClient: mainnetLiteClient,
		TestnetWallet:     testnetWallet,
		MainnetWallet:     mainnetWallet,
		Timeout:           30 * time.Second,
	})

	tonApiRepo := ton.NewTonApiRepo(tonapiClient, 30*time.Second)

	walletServiceRepo := walletservice.New(tonApiRepo, walletRepo)

	// -------------------------------- Handlers -----------------------------------------------

	walletHandler := handler.WalletHandler{
		WalletServiceRepo: walletServiceRepo,
	}

	userHandler := handler.UserHandler{
		UserService:         userServiceRepo,
		WithdrawUserService: withdrawUserRepo,
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

	go tonutil.ListenDeposits(ctx, testnetLiteApi, userRepo)
	//go tonutil.ListenDeposits(ctx, streamingApi, tonapiClient, userRepo)

	app := fiber.New(fiber.Config{
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})

	app.Use(logger.New())

	app.Get("/ping", cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}), func(c *fiber.Ctx) error {
		return c.SendString("pong")
	})

	app.Get("/favicon.ico", cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	api := app.Group("/api")
	walletApi := api.Group("/wallet")
	userApi := api.Group("/user")
	nftCollectionApi := api.Group("/nft-collection", cors.New(cors.Config{
		AllowOrigins: "https://rom6n.github.io",
		AllowMethods: "POST,GET",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))
	nftItemApi := api.Group("/nft-item", cors.New(cors.Config{
		AllowOrigins: "https://rom6n.github.io",
		AllowMethods: "POST,GET",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))
	marketApi := api.Group("/market", cors.New(cors.Config{
		AllowOrigins: "https://rom6n.github.io",
		AllowMethods: "GET,POST",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	walletApi.Get("/get-wallet-data", cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}), walletHandler.GetWalletData())
	
	walletApi.Post("/refresh-wallet-nft-items", cors.New(cors.Config{
		AllowOrigins: "https://rom6n.github.io",
		AllowMethods: "POST",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}), walletHandler.RefreshWalletNftItems())

	marketApi.Post("/deploy", marketplaceHandler.DeployMarketContract())
	marketApi.Post("/deposit", marketplaceHandler.DepositMarket())
	marketApi.Post("/withdraw", marketplaceHandler.WithdrawTonFromMarketContract())

	userApi.Get("/:id", cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}), userHandler.GetUserData())

	userApi.Get("/nft-collections/:id", cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}), userHandler.GetUserNftCollections())

	userApi.Get("/nft-items/:id", cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}), userHandler.GetUserNftItems())

	userApi.Post("/withdraw/:id", cors.New(cors.Config{
		AllowOrigins: "https://rom6n.github.io",
		AllowMethods: "POST",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}), userHandler.WithdrawUserTON())

	nftCollectionApi.Post("/deploy", nftCollectionHandler.DeployNftCollection())              // В будущем поменять на POST
	nftCollectionApi.Post("/withdraw/:address", nftCollectionHandler.WithdrawNftCollection()) // В будущем поменять на POST

	nftItemApi.Post("/mint", nftItemHandler.MintNftItem())                  // В будущем поменять на POST
	nftItemApi.Post("/withdraw/:address", nftItemHandler.WithdrawNftItem()) // В будущем поменять на POST

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
