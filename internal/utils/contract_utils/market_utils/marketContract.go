package marketutils

import (
	"crypto/ed25519"
	"encoding/hex"
	"log"
	"math/big"
	"os"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func GetMarketplaceContractCode() *cell.Cell {
	hexStr := os.Getenv("MARKETPLACE_CONTRACT_CODE")
	if hexStr == "" {
		log.Fatalln("Add marketplace contract code to .env")
	}

	boc, err := hex.DecodeString(hexStr)
	if err != nil {
		log.Fatalln("Error decode marketplace contract hex code")
	}

	code, err := cell.FromBOC(boc)
	if err != nil {
		log.Fatalln("Error decode BOC marketplace contract code")
	}

	return code
}

func GetMarketplaceContractAddress() *address.Address {
	marketplaceContractAddress := os.Getenv("MARKETPLACE_CONTRACT_ADDRESS")
	if marketplaceContractAddress == "" {
		log.Fatalln("Market contract address is not in .env")
	}
	return address.MustParseAddr(marketplaceContractAddress)
}

func GetMarketplaceContractDeployData(seqno, subwallet int32, publicKey []byte) *cell.Cell {
	return cell.BeginCell().
		MustStoreUInt(uint64(seqno), 32).
		MustStoreUInt(uint64(subwallet), 32).
		MustStoreBinarySnake(publicKey).
		EndCell()
}

func PackMessageToMarketplaceContract(privateKey ed25519.PrivateKey, validUntil int64, seqno *big.Int, mode uint64, msgToSend *cell.Cell) *cell.Cell {
	msgSigned := cell.BeginCell().
		MustStoreUInt(uint64(1947320581), 32).
		MustStoreUInt(uint64(validUntil), 64).
		MustStoreUInt(seqno.Uint64(), 32).
		MustStoreUInt(mode, 8).
		MustStoreRef(msgToSend).
		EndCell().
		Sign(privateKey)

	msg := cell.BeginCell().
		MustStoreBinarySnake(msgSigned).
		MustStoreUInt(uint64(1947320581), 32).
		MustStoreUInt(uint64(validUntil), 64).
		MustStoreUInt(seqno.Uint64(), 32).
		MustStoreUInt(mode, 8).
		MustStoreRef(msgToSend).
		EndCell()

	return msg
}
