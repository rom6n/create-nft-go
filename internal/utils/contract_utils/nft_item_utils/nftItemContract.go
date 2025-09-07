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
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
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

func PackChangeOwnerMsg(newOwner *address.Address, walletAddress *address.Address, nftItemAddress *address.Address) *tlb.InternalMessage {
	fwdMsg := cell.BeginCell().
		MustStoreUInt(0, 32).
		MustStoreStringSnake("NFT withdraw").
		EndCell()

	return &tlb.InternalMessage{
		Bounce:  true,
		DstAddr: nftItemAddress,
		Amount:  tlb.MustFromTON("0.03"), // 0.03 TON for nft item contract
		Body: cell.BeginCell().
			MustStoreUInt(0x5fcc3d14, 32). // Transfer OP-code
			MustStoreUInt(93784, 64).      // random query id
			MustStoreAddr(newOwner).
			MustStoreAddr(walletAddress).
			MustStoreInt(0, 1).
			MustStoreCoins(10000000). // 0.01 TON for forward amount
			MustStoreInt(1, 1).
			MustStoreRef(fwdMsg).
			EndCell(),
	}
}
