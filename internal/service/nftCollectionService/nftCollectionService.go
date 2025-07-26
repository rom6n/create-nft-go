package nftcollectionservice

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"log"
	"time"

	nftcollection "github.com/rom6n/create-nft-go/internal/domain/nftCollection"
	"github.com/rom6n/create-nft-go/internal/util/tonutil"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type NftCollectionServiceRepository interface {
	DeployNftCollection(ctx context.Context, mintCfg nftcollection.MintCollectionCfg) (nftcollection.NftCollection, error)
	DeployMarketplaceContract(ctx context.Context) error
}

type NftCollectionServiceRepo struct {
	NftCollectionRepo nftcollection.NftCollectionRepository
	PrivateKey        ed25519.PrivateKey
}

func New(NftCollectionRepository nftcollection.NftCollectionRepository, privateKey ed25519.PrivateKey) NftCollectionServiceRepository {
	return &NftCollectionServiceRepo{
		NftCollectionRepository,
		privateKey,
	}
}

func (v *NftCollectionServiceRepo) DeployMarketplaceContract(ctx context.Context) error {
	svcCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	client := liteclient.NewConnectionPool()
	if clientErr := client.AddConnectionsFromConfigUrl(svcCtx, "https://ton-blockchain.github.io/testnet-global.config.json"); clientErr != nil {
		log.Printf("Error connectiong to liteserver: %v\n", clientErr)
		return clientErr
	}

	api := ton.NewAPIClient(client).WithRetry()

	walletV4, walletErr := tonutil.GetTestWallet(api)
	if walletErr != nil {
		return walletErr
	}

	msgBody := cell.BeginCell().EndCell()
	deployedAddr, _, _, deployErr := walletV4.DeployContractWaitTransaction(
		svcCtx,
		tlb.MustFromTON("0.05"),
		msgBody,
		tonutil.GetMarketContractCode(),
		tonutil.GetMarketContractDeployData(0, 1947320581, []byte(ed25519.PublicKey(v.PrivateKey))),
	)
	if deployErr != nil {
		log.Printf("Error deploy market contract: %v\n", deployErr)
		return deployErr
	}

	log.Printf("Market contract deployed at address: %v\n", deployedAddr)

	return nil
}

func (v *NftCollectionServiceRepo) DeployNftCollection(ctx context.Context, mintCfg nftcollection.MintCollectionCfg) (nftcollection.NftCollection, error) {
	svcCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// TESTNET !!!!!!!!!!!!!
	client := liteclient.NewConnectionPool()
	if connErr := client.AddConnectionsFromConfigUrl(svcCtx, "https://ton-blockchain.github.io/testnet-global.config.json"); connErr != nil {
		return nftcollection.NftCollection{}, connErr
	}

	api := ton.NewAPIClient(client).WithRetry()
	apiCtx := client.StickyContext(context.Background())
	block, chainErr := api.CurrentMasterchainInfo(apiCtx)
	if chainErr != nil {
		return nftcollection.NftCollection{}, chainErr
	}

	res, methodErr := api.WaitForBlock(block.SeqNo).RunGetMethod(apiCtx, block, address.MustParseAddr("kQA64pUp9DBPh8TqOj9PS6fstMTdiuoNxcpW_R3dm3G4jfOX"), "seqno")
	if methodErr != nil {
		return nftcollection.NftCollection{}, methodErr
	}

	seqno := res.MustInt(0)

	log.Printf("Current seqno = %d", seqno)

	forwardMessage := cell.BeginCell().EndCell()
	var mode uint64 = 1

	dataSign := cell.BeginCell().
		MustStoreUInt(mode, 1).
		MustStoreRef(forwardMessage).
		EndCell().Sign(v.PrivateKey)

	subwalletId := 11111
	validUntil := time.Now().Add(1 * time.Minute).Unix()

	msgData := cell.BeginCell().
		MustStoreBinarySnake(dataSign).
		MustStoreUInt(uint64(subwalletId), 32).
		MustStoreUInt(uint64(validUntil), 32).
		MustStoreUInt(seqno.Uint64(), 32).
		MustStoreUInt(mode, 8).
		MustStoreRef(forwardMessage).
		EndCell()

	msg := &tlb.ExternalMessage{
		DstAddr: address.MustParseAddr("kQBL2_3lMiyywU17g-or8N7v9hDmPCpttzBPE2isF2GTziky"),
		Body:    msgData,
	}

	log.Println("Sending external message with hash:", hex.EncodeToString(msg.NormalizedHash()))

	msgErr := api.SendExternalMessage(apiCtx, msg)
	if msgErr != nil {
		// FYI: it can fail if not enough balance on contract
		return nftcollection.NftCollection{}, msgErr
	}

	return nftcollection.NftCollection{}, nil
}
