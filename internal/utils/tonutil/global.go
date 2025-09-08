package tonutil

import (
	"context"
	"crypto/ed25519"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/rom6n/create-nft-go/internal/domain/user"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

func GetTestnetWallet(api ton.APIClientWrapped) *wallet.Wallet {
	seedStr := os.Getenv("TEST_WALLET_SEED")
	if seedStr == "" {
		log.Fatalln("TEST WALLET SEED must be set")
	}

	seed := strings.Split(seedStr, " ")

	w, seedErr := wallet.FromSeed(api, seed, wallet.V4R2)
	if seedErr != nil {
		log.Fatalf("error creating testnet wallet: %v", seedErr)
	}

	return w
}

func GetMainnetWallet(api ton.APIClientWrapped) *wallet.Wallet {
	seedStr := os.Getenv("MAIN_WALLET_SEED")
	if seedStr == "" {
		log.Fatalln("MAIN WALLET SEED must be set")
	}

	seed := strings.Split(seedStr, " ")

	w, seedErr := wallet.FromSeed(api, seed, wallet.V4R2)
	if seedErr != nil {
		log.Fatalf("error creating mainnet wallet: %v", seedErr)
	}

	return w
}

func GetTestnetLiteClient(ctx context.Context) (*liteclient.ConnectionPool, ton.APIClientWrapped) {
	client := liteclient.NewConnectionPool()
	if err := client.AddConnectionsFromConfigUrl(ctx, "https://ton-blockchain.github.io/testnet-global.config.json"); err != nil {
		log.Fatalf("Error add connect to testnet liteclient: %v\n", err)
	}

	api := ton.NewAPIClient(client, ton.ProofCheckPolicyFast).WithRetry()
	return client, api
}

func GetMainnetLiteClient(ctx context.Context) (*liteclient.ConnectionPool, ton.APIClientWrapped) {
	client := liteclient.NewConnectionPool()
	if err := client.AddConnectionsFromConfigUrl(ctx, "https://ton-blockchain.github.io/global.config.json"); err != nil {
		log.Fatalf("Error add connect to mainnet liteclient: %v\n", err)
	}

	api := ton.NewAPIClient(client, ton.ProofCheckPolicyFast).WithRetry()
	return client, api
}

func GetPrivateKey() ed25519.PrivateKey {
	seedStr := os.Getenv("PRIVATE_KEY_SEED")
	if seedStr == "" {
		log.Fatalln("Add PRIVATE_KEY_SEED to .env")
	}

	if len(seedStr) != ed25519.SeedSize {
		log.Fatalf("ED25519 SEED SIZE MUST BE: %v; HAVE: %v", ed25519.SeedSize, len(seedStr))
	}

	seed := make([]byte, 32)
	copy(seed, []byte(seedStr))

	private := ed25519.NewKeyFromSeed(seed)

	return private
}

//func GetStreamingApi() *tonapi.StreamingAPI {
//	token := os.Getenv("TONAPI_TOKEN")
//	if token == "" {
//		log.Fatal("Error. Add TonApi token to env.")
//	}
//	return tonapi.NewStreamingAPI(tonapi.WithStreamingToken(token))
//}
//
//func GetTestnetStreamingApi() *tonapi.StreamingAPI {
//	token := os.Getenv("TONAPI_TOKEN")
//	if token == "" {
//		log.Fatal("Error. Add TonApi token to env.")
//	}
//	return tonapi.NewStreamingAPI(tonapi.WithStreamingEndpoint(tonapi.TestnetTonApiURL), tonapi.WithStreamingToken(token))
//}

func ListenDeposits(ctx context.Context, api ton.APIClientWrapped, userRepo user.UserRepository) {
	master, err := api.CurrentMasterchainInfo(context.Background()) // we fetch block just to trigger chain proof check
	if err != nil {
		log.Fatalln("get masterchain info err: ", err.Error())
		return
	}

	treasuryAddress := address.MustParseAddr("kQDU46qYz4rHAJhszrW9w6imF8p4Cw5dS1GpPTcJ9vqNSjQa")

	acc, err := api.GetAccount(context.Background(), master, treasuryAddress)
	if err != nil {
		log.Fatalln("get masterchain info err: ", err.Error())
		return
	}

	log.Printf("Account is active: %v", acc.IsActive)

	lastProcessedLT := acc.LastTxLT
	transactions := make(chan *tlb.Transaction)

	log.Printf("ListenDeposits запущен!\n")
	go api.SubscribeOnTransactions(ctx, treasuryAddress, lastProcessedLT, transactions)
	for tx := range transactions {
		if tx.IO.In != nil && tx.IO.In.MsgType == tlb.MsgTypeInternal {
			ti := tx.IO.In.AsInternal()

			if dsc, ok := tx.Description.(tlb.TransactionDescriptionOrdinary); ok && dsc.BouncePhase != nil {
				if _, ok = dsc.BouncePhase.Phase.(tlb.BouncePhaseOk); ok {
					// transaction was bounced, and coins were returned to sender
					// this can happen mostly on custom contracts
					continue
				}
			}

			if ti.Amount.Nano().Sign() > 0 && ti.Comment() != "" {
				userID, parseErr := strconv.ParseInt(ti.Comment(), 0, 64)
				if parseErr != nil {
					log.Printf("Deposits listener: Error parsing user id to int64: %v\n", parseErr)
					continue
				}

				user, getErr := userRepo.GetUserByID(ctx, userID)
				if getErr != nil {
					log.Printf("Deposits listener: Error getting user: %v\n", getErr)
					continue
				}

				receivedNanoTon, parseErr := strconv.ParseUint(ti.Amount.Nano().String(), 0, 64)
				if parseErr != nil {
					log.Printf("Deposits listener: Error parsing received nano ton to uint64: %v\n", parseErr)
					continue
				}

				updErr := userRepo.UpdateUserBalance(ctx, user.UUID, user.NanoTon+receivedNanoTon)
				if updErr != nil {
					log.Printf("Deposits listener: Error updating user's balance: %v\n", updErr)
				}
			}
		}

	}

}
