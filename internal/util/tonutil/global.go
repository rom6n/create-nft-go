package tonutil

import (
	"context"
	"crypto/ed25519"
	"log"
	"os"
	"strings"

	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

func GetTestWallet(api ton.APIClientWrapped) (*wallet.Wallet, error) {
	seedStr := os.Getenv("TEST_WALLET_SEED")
	if seedStr == "" {
		log.Fatalln("TEST WALLET SEED must be set")
	}

	seed := strings.Split(seedStr, " ")

	w, seedErr := wallet.FromSeed(api, seed, wallet.V4R2)
	if seedErr != nil {
		return &wallet.Wallet{}, seedErr
	}

	return w, nil
}

func GetLiteClient(ctx context.Context) (*liteclient.ConnectionPool, ton.APIClientWrapped) {
	client := liteclient.NewConnectionPool()
	if err := client.AddConnectionsFromConfigUrl(ctx, "https://ton-blockchain.github.io/testnet-global.config.json"); err != nil {
		log.Fatalf("Error add connect to liteclient: %v\n", err)
	}

	api := ton.NewAPIClient(client)

	return client, api
}

func GetTestPrivateKey() ed25519.PrivateKey {
	seedStr := os.Getenv("TEST_PRIVATE_KEY_SEED")
	if seedStr == "" {
		log.Fatalln("TEST_PRIVATE_KEY_SEED must be set")
	}

	if len(seedStr) != ed25519.SeedSize {
		log.Fatalf("ED25519 SEED SIZE MUST BE: %v; HAVE: %v", ed25519.SeedSize, len(seedStr))
	}

	seed := make([]byte, 32)
	copy(seed, []byte(seedStr))

	private := ed25519.NewKeyFromSeed([]byte(seed))

	return private
}
