package deploynftcollection

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"log"
	"time"

	nftcollection "github.com/rom6n/create-nft-go/internal/domain/nftCollection"
	"github.com/rom6n/create-nft-go/internal/domain/user"
	userservice "github.com/rom6n/create-nft-go/internal/service/userService"
	contractutils "github.com/rom6n/create-nft-go/internal/util/contractUtils"
	"github.com/rom6n/create-nft-go/internal/util/contractUtils/marketutils"
	"github.com/rom6n/create-nft-go/internal/util/contractUtils/nftcollectionutils"
	"github.com/rom6n/create-nft-go/internal/util/contractUtils/nftitemutils"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
)

type DeployNftCollectionServiceRepository interface {
	DeployNftCollection(ctx context.Context, deployCfg nftcollection.DeployCollectionCfg, ownerID int64) (*nftcollection.NftCollection, error)
}

type deployNftCollectionServiceRepo struct {
	NftCollectionRepo nftcollection.NftCollectionRepository
	UserRepo          userservice.UserServiceRepository
	PrivateKey        ed25519.PrivateKey
	LiteClient        *liteclient.ConnectionPool
	LiteclientApi     ton.APIClientWrapped
}

func New(NftCollectionRepo nftcollection.NftCollectionRepository, userCollectionRepo user.UserRepository, privateKey ed25519.PrivateKey, liteClient *liteclient.ConnectionPool, liteclientApi ton.APIClientWrapped) DeployNftCollectionServiceRepository {
	return &deployNftCollectionServiceRepo{
		NftCollectionRepo,
		userCollectionRepo,
		privateKey,
		liteClient,
		liteclientApi,
	}
}

func (v *deployNftCollectionServiceRepo) DeployNftCollection(ctx context.Context, deployCfg nftcollection.DeployCollectionCfg, ownerID int64) (*nftcollection.NftCollection, error) {
	svcCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// TESTNET !!!!!!!!!!!!!
	marketAddress := "EQC4erFNnXJK_keR6eJtoG_70f6ygf1hnh1VYDOHZ0W4oqVh"
	client := v.LiteClient
	api := v.LiteclientApi
	apiCtx := client.StickyContext(svcCtx)

	//
	//
	//
	// Должен уменьшать баланс пользователя на цену деплоя коллекции
	//
	//
	//
	//

	block, chainErr := api.CurrentMasterchainInfo(apiCtx)
	if chainErr != nil {
		return &nftcollection.NftCollection{}, chainErr
	}

	response, methodErr := api.WaitForBlock(block.SeqNo).RunGetMethod(apiCtx, block, address.MustParseAddr(marketAddress), "seqno")
	if methodErr != nil {
		return &nftcollection.NftCollection{}, methodErr
	}

	seqno := response.MustInt(0)
	log.Printf("Current seqno = %d", seqno)

	ownerAccount, userErr := v.UserRepo.GetUserByID(svcCtx, ownerID)
	if userErr != nil {
		return &nftcollection.NftCollection{}, userErr
	}

	codeCell, codeErr := nftcollectionutils.GetNftCollectionContractCode()
	if codeErr != nil {
		return &nftcollection.NftCollection{}, codeErr
	}

	eachNftItemCode, nftCodeErr := nftitemutils.GetNftItemCode()
	if nftCodeErr != nil {
		return &nftcollection.NftCollection{}, nftCodeErr
	}

	content := nftcollectionutils.PackOffchainContentForNftCollection(deployCfg.CollectionContent, deployCfg.CommonContent)

	royaltyParams := nftcollectionutils.PackNftCollectionRoyaltyParams(deployCfg.RoyaltyDividend, deployCfg.RoyaltyDivisor, deployCfg.OwnerAddress)

	if deployCfg.OwnerAddress == "" {
		deployCfg.OwnerAddress = marketAddress
	}

	dataCell := nftcollectionutils.PackNftCollectionData(deployCfg.OwnerAddress, content, eachNftItemCode, royaltyParams)

	stateInit := contractutils.PackStateInit(codeCell, dataCell)

	toAddress := contractutils.CalculateAddress(0, stateInit)

	deployMsg := contractutils.PackDeployMessage(toAddress, stateInit)

	validUntil := time.Now().Add(1 * time.Minute).Unix()

	msgData := marketutils.PackMessageToMarketContract(v.PrivateKey, validUntil, seqno, 0, deployMsg)

	var nftCollectionMetadata nftcollection.NftCollectionMetadata

	metadataErr := nftcollectionutils.GetNftCollectionMetadataByLink(deployCfg.CollectionContent, nftCollectionMetadata)
	if metadataErr != nil {
		return &nftcollection.NftCollection{}, metadataErr
	}

	nftCollection := nftcollection.New(toAddress.String(), ownerAccount.UUID, &nftCollectionMetadata)

	if deployCfg.OwnerAddress == marketAddress {
		createErr := v.NftCollectionRepo.CreateCollection(svcCtx, nftCollection, ownerAccount.UUID)
		if createErr != nil {
			return &nftcollection.NftCollection{}, createErr
		}
	}

	msg := &tlb.ExternalMessage{
		DstAddr: address.MustParseAddr(marketAddress),
		Body:    msgData,
	}
	log.Println("Sending external message with hash:", hex.EncodeToString(msg.NormalizedHash()))

	msgErr := api.SendExternalMessage(apiCtx, msg)

	if msgErr != nil {
		// FYI: it can fail if not enough balance on contract
		if deployCfg.OwnerAddress == marketAddress {
			for i := 0; i < 10; i++ {
				deleteErr := v.NftCollectionRepo.DeleteCollection(svcCtx, toAddress.String())
				if deleteErr == nil {
					break
				}
				log.Printf("‼️ Cant delete NFT Collection: %v, on error from DB, TRY: %v\n", toAddress.String(), i)
				time.Sleep(1 * time.Second)
			}
		}
		return &nftcollection.NftCollection{}, msgErr
	}

	log.Printf("NFT COLLECTION DEPLOYED AT ADDRESS: %v\n", toAddress.String())

	return nftCollection, nil
}
