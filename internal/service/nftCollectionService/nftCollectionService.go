package nftcollectionservice

import (
	"context"
	"encoding/hex"
	"log"
	"time"

	nftcollection "github.com/rom6n/create-nft-go/internal/domain/nftCollection"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type NftCollectionServiceRepository interface {
	MintNftCollection(ctx context.Context, mintCfg nftcollection.MintCollectionCfg) (nftcollection.NftCollection, error)
}

type NftCollectionServiceRepo struct {
	NftCollectionRepo nftcollection.NftCollectionRepository
}

func (v *NftCollectionServiceRepo) MintNftCollection(ctx context.Context, mintCfg nftcollection.MintCollectionCfg) (nftcollection.NftCollection, error) {
	svcCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	client := liteclient.NewConnectionPool()
	if connErr := client.AddConnection(svcCtx, "135.181.140.212:13206", "K0t3+IWLOXHYMvMcrGZDPs+pn58a17LFbnXoQkKc2xw="); connErr != nil {
		return nftcollection.NftCollection{}, connErr
	}

	api := ton.NewAPIClient(client).WithRetry()
	apiCtx := client.StickyContext(context.Background())
	block, chainErr := api.CurrentMasterchainInfo(apiCtx)
	if chainErr != nil {
		return nftcollection.NftCollection{}, chainErr
	}

	res, methodErr := api.WaitForBlock(block.SeqNo).RunGetMethod(apiCtx, block, address.MustParseAddr("kQBL2_3lMiyywU17g-or8N7v9hDmPCpttzBPE2isF2GTziky"), "get_total")
	if methodErr != nil {
		return nftcollection.NftCollection{}, methodErr
	}

	seqno := res.MustInt(0)
	total := res.MustInt(1)

	log.Printf("Current seqno = %d and total = %d", seqno, total)

	data := cell.BeginCell().
		MustStoreBigInt(seqno, 64).
		MustStoreUInt(1, 16). // add 1 to total
		EndCell()

	msg := &tlb.ExternalMessage{
		DstAddr: address.MustParseAddr("kQBL2_3lMiyywU17g-or8N7v9hDmPCpttzBPE2isF2GTziky"),
		Body:    data,
	}

	log.Println("Sending external message with hash:", hex.EncodeToString(msg.NormalizedHash()))

	msgErr := api.SendExternalMessage(ctx, msg)
	if msgErr != nil {
		// FYI: it can fail if not enough balance on contract
		return nftcollection.NftCollection{}, msgErr
	}

	return nftcollection.NftCollection{}, nil
}
