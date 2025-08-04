package nftitemutils

import (
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/goccy/go-json"
	nftitem "github.com/rom6n/create-nft-go/internal/domain/nft_item"
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

func GetNftItemOffchainMetadata(link string) (*nftitem.NftItemMetadata, error) {
	body, metadataErr := http.Get(link)

	if metadataErr != nil {
		return nil, metadataErr
	}

	rawNftItemMetadata, err := io.ReadAll(body.Body)
	defer body.Body.Close()

	if err != nil {
		return nil, fmt.Errorf("reading nft item metadata body failed: %w", err)
	}

	var parseTo nftitem.NftItemMetadata

	if unmarshErr := json.Unmarshal(rawNftItemMetadata, &parseTo); unmarshErr != nil {
		return nil, unmarshErr
	}

	return &parseTo, nil
}
