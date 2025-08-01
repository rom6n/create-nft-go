package nftcollectionutils

import (
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/goccy/go-json"
	nftcollection "github.com/rom6n/create-nft-go/internal/domain/nft_collection"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func GetNftCollectionContractCode() *cell.Cell {
	hexStr := os.Getenv("NFT_COLLECTION_CONTRACT_CODE")
	if hexStr == "" {
		log.Fatalln("Nft collection contract code is not in .env")
	}

	boc, err := hex.DecodeString(hexStr)
	if err != nil {
		log.Fatalf("Error decoding nft collection hex code: %v\n", err)
	}

	code, err := cell.FromBOC(boc)
	if err != nil {
		log.Fatalf("Error decoding from BOC nft collection code: %v\n", err)
	}

	return code
}

func GetNftCollectionMetadataByLink(link string) (nftcollection.NftCollectionMetadata, error) {
	body, metadataErr := http.Get(link)
	if metadataErr != nil {
		return nftcollection.NftCollectionMetadata{}, metadataErr
	}
	defer body.Body.Close()

	rawNftCollectionMetadata, err := io.ReadAll(body.Body)
	if err != nil {
		return nftcollection.NftCollectionMetadata{}, fmt.Errorf("reading body failed: %w", err)
	}

	var parseTo nftcollection.NftCollectionMetadata

	if unmarshErr := json.Unmarshal(rawNftCollectionMetadata, &parseTo); unmarshErr != nil {
		return nftcollection.NftCollectionMetadata{}, unmarshErr
	}

	return parseTo, nil
}

func PackOffchainContentForNftCollection(collectionContent string, commonContent string) *cell.Cell {
	collectionCont := cell.BeginCell().
		MustStoreUInt(1, 8).
		MustStoreStringSnake(collectionContent).
		EndCell()

	commonCont := cell.BeginCell().
		MustStoreStringSnake(commonContent).
		EndCell()

	content := cell.BeginCell().
		MustStoreRef(collectionCont).
		MustStoreRef(commonCont).
		EndCell()

	return content
}

func PackNftCollectionRoyaltyParams(royaltyDividend uint16, royaltyDivisor uint16, royaltyAddress string) *cell.Cell {
	return cell.BeginCell().
		MustStoreUInt(uint64(royaltyDividend), 16).
		MustStoreUInt(uint64(royaltyDivisor), 16).
		MustStoreAddr(address.MustParseAddr(royaltyAddress)).
		EndCell()
}

func PackNftCollectionData(ownerAddress string, content *cell.Cell, nftItemCode *cell.Cell, royaltyParams *cell.Cell) *cell.Cell {
	return cell.BeginCell().
		MustStoreAddr(address.MustParseAddr(ownerAddress)).
		MustStoreUInt(1, 64).
		MustStoreRef(content).
		MustStoreRef(nftItemCode).
		MustStoreRef(royaltyParams).
		EndCell()
}
