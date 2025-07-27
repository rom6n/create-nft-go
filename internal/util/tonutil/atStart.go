package tonutil

import (
	"crypto/ed25519"
	"log"
	"os"
	"strings"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"
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

	/*text2sign := "hello"

	log.Printf("PRIVATE KEY SIZE: %v\nMUST BE: %v\n", len(private), ed25519.PrivateKeySize)

	publicbyte := []byte(private.Public().(ed25519.PublicKey))
	public := ed25519.PublicKey(publicbyte)

	log.Printf("PUBLIC KEY SIZE\nSIZE PUBLIC KEY: %v\nSIZE PUBLIC BYTE KEY: %v\nMUST BE: %v\n", len(public), len(publicbyte), ed25519.PublicKeySize)

	signature := ed25519.Sign(private, []byte(text2sign))
	log.Printf("SIGNATURE SIZE: %v\nMUST BE: %v", len(signature), ed25519.SignatureSize)
	isOK := ed25519.Verify(publicbyte, []byte(text2sign), signature)
	log.Fatalf("Is OK: %v", isOK)*/

	return private
}

func CalculateAddress(workchain int32, stateInit *cell.Cell) *address.Address {
	address := address.NewAddress(4, 0, stateInit.Hash())
	/*addressSlice := cell.BeginCell().
	MustStoreUInt(4, 3).
	MustStoreInt(int64(workchain), 8).
	MustStoreSlice(stateInit.Hash(), 256).
	EndCell().BeginParse()*/
	return address
}
