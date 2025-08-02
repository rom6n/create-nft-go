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
	nftitem "github.com/rom6n/create-nft-go/internal/domain/nft_item"
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

func GetNftCollectionMetadataByLink(link string) (*nftcollection.NftCollectionMetadata, error) {
	body, metadataErr := http.Get(link)
	nilNftCollectionMetadata := &nftcollection.NftCollectionMetadata{}

	if metadataErr != nil {
		return nilNftCollectionMetadata, metadataErr
	}

	rawNftCollectionMetadata, err := io.ReadAll(body.Body)
	defer body.Body.Close()

	if err != nil {
		return nilNftCollectionMetadata, fmt.Errorf("reading body failed: %w", err)
	}

	var parseTo nftcollection.NftCollectionMetadata

	if unmarshErr := json.Unmarshal(rawNftCollectionMetadata, &parseTo); unmarshErr != nil {
		return nilNftCollectionMetadata, unmarshErr
	}

	return &parseTo, nil
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

func PackDeployNftItemMessage(nftCcollectionAddress *address.Address, nextItemIndex uint64, cfg nftitem.DeployNftItemCfg) *cell.Cell {
	fwdMsg := cell.BeginCell().
		MustStoreUInt(0, 32).
		MustStoreStringSnake(cfg.ForwardMessage).
		EndCell()

	content := cell.BeginCell().MustStoreStringSnake(cfg.Content).EndCell()

	initContent := cell.BeginCell().
		MustStoreAddr(cfg.OwnerAddress).
		MustStoreRef(content).
		MustStoreCoins(cfg.ForwardAmount).
		MustStoreInt(1, 1).
		MustStoreRef(fwdMsg).
		EndCell()

	return cell.BeginCell().
		MustStoreUInt(0x10, 6).
		MustStoreAddr(nftCcollectionAddress).
		MustStoreCoins(120000000).
		MustStoreUInt(0, 1+4+4+64+32+1+1).
		MustStoreUInt(1, 32).
		MustStoreUInt(0, 64).
		MustStoreUInt(nextItemIndex, 64).
		MustStoreCoins(110000000).
		MustStoreRef(initContent).
		EndCell()
}
