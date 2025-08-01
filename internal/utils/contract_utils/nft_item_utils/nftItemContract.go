package nftitemutils

import (
	"encoding/hex"
	"log"
	"os"

	"github.com/xssnick/tonutils-go/tvm/cell"
)

func GetNftItemContractCode() *cell.Cell {
	codeHex := os.Getenv("NFT_ITEM_CONTRACT_CODE")
	if codeHex == "" {
		log.Fatalln("Nft item contract code is not in .env")
	}

	boc, err := hex.DecodeString(codeHex)
	if err != nil {
		log.Fatalf("Error decoding nft item hex code: %v\n", err)
	}

	code, err := cell.FromBOC(boc)
	if err != nil {
		log.Fatalf("Error decoding from BOC nft item code: %v\n", err)
	}

	return code
}
