package nftcollectionservice

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"log"
	"time"

	nftcollection "github.com/rom6n/create-nft-go/internal/domain/nftCollection"
	"github.com/rom6n/create-nft-go/internal/domain/user"
	userservice "github.com/rom6n/create-nft-go/internal/service/userService"
	"github.com/rom6n/create-nft-go/internal/util/tonutil"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type NftCollectionServiceRepository interface {
	DeployNftCollection(ctx context.Context, mintCfg nftcollection.DeployCollectionCfg, ownerID int64) (*nftcollection.NftCollection, error)
	DeployMarketplaceContract(ctx context.Context) error
}

type nftCollectionServiceRepo struct {
	NftCollectionRepo nftcollection.NftCollectionRepository
	UserRepo          userservice.UserServiceRepository
	PrivateKey        ed25519.PrivateKey
}

func New(NftCollectionRepo nftcollection.NftCollectionRepository, userCollectionRepo user.UserRepository, privateKey ed25519.PrivateKey) NftCollectionServiceRepository {
	return &nftCollectionServiceRepo{
		NftCollectionRepo,
		userCollectionRepo,
		privateKey,
	}
}

func (v *nftCollectionServiceRepo) DeployMarketplaceContract(ctx context.Context) error {
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
		tlb.MustFromTON("0.07"),
		msgBody,
		tonutil.GetMarketContractCode(),
		tonutil.GetMarketContractDeployData(0, 1947320581, []byte(v.PrivateKey.Public().(ed25519.PublicKey))),
	)
	if deployErr != nil {
		log.Printf("Error deploy market contract: %v\n", deployErr)
		return deployErr
	}

	log.Printf("Market contract deployed at address: %v\n", deployedAddr)

	return nil
}

func (v *nftCollectionServiceRepo) DeployNftCollection(ctx context.Context, mintCfg nftcollection.DeployCollectionCfg, ownerID int64) (*nftcollection.NftCollection, error) {
	svcCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// TESTNET !!!!!!!!!!!!!
	marketAddress := "EQC4erFNnXJK_keR6eJtoG_70f6ygf1hnh1VYDOHZ0W4oqVh"
	client := liteclient.NewConnectionPool()
	if connErr := client.AddConnectionsFromConfigUrl(svcCtx, "https://ton-blockchain.github.io/testnet-global.config.json"); connErr != nil {
		return &nftcollection.NftCollection{}, connErr
	}

	api := ton.NewAPIClient(client).WithRetry()
	apiCtx := client.StickyContext(svcCtx)

	block, chainErr := api.CurrentMasterchainInfo(apiCtx)
	if chainErr != nil {
		return &nftcollection.NftCollection{}, chainErr
	}

	res, methodErr := api.WaitForBlock(block.SeqNo).RunGetMethod(apiCtx, block, address.MustParseAddr(marketAddress), "seqno")
	if methodErr != nil {
		return &nftcollection.NftCollection{}, methodErr
	}

	seqno := res.MustInt(0)
	log.Printf("Current seqno = %d", seqno)

	owner, userErr := v.UserRepo.GetUserByID(svcCtx, ownerID)
	if userErr != nil {
		return &nftcollection.NftCollection{}, userErr
	}

	codeCell, codeErr := tonutil.GetNftCollectionContractCode()
	if codeErr != nil {
		return &nftcollection.NftCollection{}, codeErr
	}

	eachNftItemCode, nftCodeErr := tonutil.GetNftItemCode()
	if nftCodeErr != nil {
		return &nftcollection.NftCollection{}, nftCodeErr
	}

	collectionContent := cell.BeginCell().
		MustStoreUInt(1, 8).
		MustStoreStringSnake(mintCfg.CollectionContent).
		EndCell()
	commonContent := cell.BeginCell().
		MustStoreStringSnake(mintCfg.CommonContent).
		EndCell()
	content := cell.BeginCell().
		MustStoreRef(collectionContent).
		MustStoreRef(commonContent).
		EndCell()

	royaltyParams := cell.BeginCell().
		MustStoreUInt(uint64(mintCfg.RoyaltyDividend), 16).
		MustStoreUInt(uint64(mintCfg.RoyaltyDivisor), 16).
		MustStoreAddr(address.MustParseAddr(mintCfg.Owner)).
		EndCell()

	if mintCfg.Owner == "" {
		mintCfg.Owner = marketAddress
	}

	dataCell := cell.BeginCell().
		MustStoreAddr(address.MustParseAddr(mintCfg.Owner)).
		MustStoreUInt(1, 64).
		MustStoreRef(content).
		MustStoreRef(eachNftItemCode).
		MustStoreRef(royaltyParams).
		EndCell()

	stateInit := cell.BeginCell().
		MustStoreUInt(6, 5).
		MustStoreRef(codeCell).
		MustStoreRef(dataCell).
		EndCell()

	toAddress := tonutil.CalculateAddress(0, stateInit)

	forwardMessage := cell.BeginCell().
		MustStoreUInt(0x18, 6).
		MustStoreAddr(toAddress).
		MustStoreCoins(50000000).
		MustStoreUInt(4+2+1, 1+4+4+64+32+1+1+1).
		MustStoreRef(stateInit).
		MustStoreRef(cell.BeginCell().EndCell()).
		EndCell()

	var mode uint64 = 0
	subwalletId := 1947320581
	validUntil := time.Now().Add(1 * time.Minute).Unix()

	dataSigned := cell.BeginCell().
		MustStoreUInt(uint64(subwalletId), 32).
		MustStoreUInt(uint64(validUntil), 64).
		MustStoreUInt(seqno.Uint64(), 32).
		MustStoreUInt(mode, 8).
		MustStoreRef(forwardMessage).
		EndCell().
		Sign(v.PrivateKey)

	msgData := cell.BeginCell().
		MustStoreBinarySnake(dataSigned).
		MustStoreUInt(uint64(subwalletId), 32).
		MustStoreUInt(uint64(validUntil), 64).
		MustStoreUInt(seqno.Uint64(), 32).
		MustStoreUInt(mode, 8).
		MustStoreRef(forwardMessage).
		EndCell()

	msg := &tlb.ExternalMessage{
		DstAddr: address.MustParseAddr(marketAddress),
		Body:    msgData,
	}

	nftCollection := &nftcollection.NftCollection{
		Address:       toAddress.String(),
		NextItemIndex: 1,
		Owner:         owner.UUID,
		Metadata:      nftcollection.NftCollectionMetadata{},
	}

	if mintCfg.Owner == marketAddress {
		createErr := v.NftCollectionRepo.CreateCollection(svcCtx, nftCollection, owner.UUID)
		if createErr != nil {
			return &nftcollection.NftCollection{}, createErr
		}
	}

	log.Println("Sending external message with hash:", hex.EncodeToString(msg.NormalizedHash()))

	msgErr := api.SendExternalMessage(apiCtx, msg)

	if msgErr != nil {
		// FYI: it can fail if not enough balance on contract
		if mintCfg.Owner == marketAddress {
			for i := 0; i < 10; i++ {
				deleteErr := v.NftCollectionRepo.DeleteCollection(svcCtx, toAddress.String())
				if deleteErr == nil {
					break
				}
				log.Printf("‼️ Cant delete NFT Collection: %v, on error from DB, TRY: %v\n", toAddress.String(), i)
				time.Sleep(1*time.Second)
			}
		}
		return &nftcollection.NftCollection{}, msgErr
	}

	log.Printf("NFT COLLECTION DEPLOYED AT ADDRESS: %v\n", toAddress.String())

	return nftCollection, nil
}
