package nftcollectionutils

import (
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

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

func GetNftCollectionOffchainMetadata(link string) (*nftcollection.NftCollectionMetadata, error) {
	body, metadataErr := http.Get(link)

	if metadataErr != nil {
		return nil, metadataErr
	}

	rawNftCollectionMetadata, err := io.ReadAll(body.Body)
	defer body.Body.Close()

	if err != nil {
		return nil, fmt.Errorf("reading body failed: %w", err)
	}

	var parseTo nftcollection.NftCollectionMetadata

	if unmarshErr := json.Unmarshal(rawNftCollectionMetadata, &parseTo); unmarshErr != nil {
		return nil, unmarshErr
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

func PackNftCollectionRoyaltyParams(royaltyDividend uint16, royaltyDivisor uint16, royaltyAddress *address.Address) *cell.Cell {
	return cell.BeginCell().
		MustStoreUInt(uint64(royaltyDividend), 16).
		MustStoreUInt(uint64(royaltyDivisor), 16).
		MustStoreAddr(royaltyAddress).
		EndCell()
}

func PackNftCollectionData(ownerAddress *address.Address, content *cell.Cell, nftItemCode *cell.Cell, royaltyParams *cell.Cell) *cell.Cell {
	return cell.BeginCell().
		MustStoreAddr(ownerAddress).
		MustStoreUInt(1, 64).
		MustStoreRef(content).
		MustStoreRef(nftItemCode).
		MustStoreRef(royaltyParams).
		EndCell()
}

func PackDeployNftItemMessage(nftCcollectionAddress *address.Address, nextItemIndex uint64, cfg nftitem.MintNftItemCfg) *cell.Cell {
	cfg.Content = strings.TrimPrefix(cfg.Content, "https://")
	content := cell.BeginCell().MustStoreStringSnake(cfg.Content).EndCell()
	amount := uint64(70000000)

	initContent := cell.BeginCell().
		MustStoreAddr(cfg.OwnerAddress).
		MustStoreRef(content)

	if cfg.ForwardAmount >= 1 {
		initContent.MustStoreCoins(cfg.ForwardAmount)
		amount += cfg.ForwardAmount

		if cfg.ForwardMessage != "" {
			fwdMsg := cell.BeginCell().
				MustStoreUInt(0, 32).
				MustStoreStringSnake(cfg.ForwardMessage).
				EndCell()

			initContent.
				MustStoreInt(1, 1).
				MustStoreRef(fwdMsg)
		}
	}

	return cell.BeginCell().
		MustStoreUInt(0x10, 6).
		MustStoreAddr(nftCcollectionAddress).
		MustStoreCoins(amount).
		MustStoreUInt(0, 1+4+4+64+32+1+1).
		MustStoreUInt(1, 32).
		MustStoreUInt(0, 64).
		MustStoreUInt(nextItemIndex, 64).
		MustStoreCoins(amount - 10000000). // -0.01 TON
		MustStoreRef(initContent.EndCell()).
		EndCell()
}

func PackChangeOwnerMsg(newOwner *address.Address, nftCollectionAddress *address.Address) *cell.Cell {
	return cell.BeginCell().
		MustStoreUInt(0x10, 6).
		MustStoreAddr(nftCollectionAddress).
		MustStoreCoins(10000000).
		MustStoreUInt(0, 1+4+4+64+32+1+1).
		MustStoreUInt(3, 32).
		MustStoreUInt(5235, 64). 
		MustStoreAddr(newOwner).
		EndCell()
}