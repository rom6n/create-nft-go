package deploynftcollection

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	nftcollection "github.com/rom6n/create-nft-go/internal/domain/nft_collection"
	"github.com/rom6n/create-nft-go/internal/domain/user"
	generalcontractutils "github.com/rom6n/create-nft-go/internal/utils/contract_utils/general_contract_utils"
	marketutils "github.com/rom6n/create-nft-go/internal/utils/contract_utils/market_utils"
	nftcollectionutils "github.com/rom6n/create-nft-go/internal/utils/contract_utils/nft_collection_utils"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type DeployNftCollectionServiceRepository interface {
	DeployNftCollection(ctx context.Context, deployCfg nftcollection.DeployCollectionCfg, ownerID int64) (*nftcollection.NftCollection, error)
}

type deployNftCollectionServiceRepo struct {
	NftCollectionRepo          nftcollection.NftCollectionRepository
	UserRepo                   user.UserRepository
	PrivateKey                 ed25519.PrivateKey
	LiteClient                 *liteclient.ConnectionPool
	LiteclientApi              ton.APIClientWrapped
	MarketplaceContractAddress *address.Address
	NftCollectionContractCode  *cell.Cell
	NftItemContractCode        *cell.Cell
}

type DeployNftCollectionServiceCfg struct {
	NftCollectionRepo          nftcollection.NftCollectionRepository
	UserRepo                   user.UserRepository
	PrivateKey                 ed25519.PrivateKey
	LiteClient                 *liteclient.ConnectionPool
	LiteclientApi              ton.APIClientWrapped
	MarketplaceContractAddress *address.Address
	NftCollectionContractCode  *cell.Cell
	NftItemContractCode        *cell.Cell
}

func New(cfg DeployNftCollectionServiceCfg) DeployNftCollectionServiceRepository {
	return &deployNftCollectionServiceRepo{
		cfg.NftCollectionRepo,
		cfg.UserRepo,
		cfg.PrivateKey,
		cfg.LiteClient,
		cfg.LiteclientApi,
		cfg.MarketplaceContractAddress,
		cfg.NftCollectionContractCode,
		cfg.NftItemContractCode,
	}
}

func (v *deployNftCollectionServiceRepo) DeployNftCollection(ctx context.Context, deployCfg nftcollection.DeployCollectionCfg, ownerID int64) (*nftcollection.NftCollection, error) {
	svcCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	marketAddress := v.MarketplaceContractAddress
	client := v.LiteClient
	api := v.LiteclientApi
	apiCtx := client.StickyContext(svcCtx)
	nanoTonsForDeploy := uint64(65000000)

	ownerAccount, userErr := v.UserRepo.GetUserByID(svcCtx, ownerID)
	if userErr != nil {
		return &nftcollection.NftCollection{}, userErr
	}

	if ownerAccount.NanoTon < uint64(nanoTonsForDeploy) { // return if not enough TON
		return &nftcollection.NftCollection{}, fmt.Errorf("not enough toncoins. Need %v nano ton more", (nanoTonsForDeploy - ownerAccount.NanoTon))
	}

	block, chainErr := api.CurrentMasterchainInfo(apiCtx)
	if chainErr != nil {
		return &nftcollection.NftCollection{}, chainErr
	}

	response, methodErr := api.WaitForBlock(block.SeqNo).RunGetMethod(apiCtx, block, marketAddress, "seqno")
	if methodErr != nil {
		return &nftcollection.NftCollection{}, methodErr
	}

	seqno := response.MustInt(0)
	log.Printf("Current seqno = %d", seqno)

	content := nftcollectionutils.PackOffchainContentForNftCollection(deployCfg.CollectionContent, deployCfg.CommonContent)

	royaltyParams := nftcollectionutils.PackNftCollectionRoyaltyParams(deployCfg.RoyaltyDividend, deployCfg.RoyaltyDivisor, deployCfg.OwnerAddress)

	if deployCfg.OwnerAddress == "" {
		deployCfg.OwnerAddress = marketAddress.String()
	}

	dataCell := nftcollectionutils.PackNftCollectionData(deployCfg.OwnerAddress, content, v.NftItemContractCode, royaltyParams)

	stateInit := generalcontractutils.PackStateInit(v.NftCollectionContractCode, dataCell)

	toAddress := generalcontractutils.CalculateAddress(0, stateInit)

	deployMsg := generalcontractutils.PackDeployMessage(toAddress, stateInit)

	validUntil := time.Now().Add(1 * time.Minute).Unix()

	msgData := marketutils.PackMessageToMarketplaceContract(v.PrivateKey, validUntil, seqno, 1, deployMsg)

	nftCollectionMetadata, metadataErr := nftcollectionutils.GetNftCollectionMetadataByLink(deployCfg.CollectionContent)
	if metadataErr != nil {
		return &nftcollection.NftCollection{}, metadataErr
	}

	nftCollection := nftcollection.New(toAddress.String(), ownerAccount.UUID, nftCollectionMetadata)

	if deployCfg.OwnerAddress == marketAddress.String() {
		createErr := v.NftCollectionRepo.CreateCollection(svcCtx, nftCollection, ownerAccount.UUID)
		if createErr != nil {
			return &nftcollection.NftCollection{}, createErr
		}
	}

	if updErr := v.UserRepo.UpdateUserBalance(svcCtx, ownerAccount.UUID, ownerAccount.NanoTon-nanoTonsForDeploy); updErr != nil { // reducing the balance before deploy. if error deleting collection
		if deployCfg.OwnerAddress == marketAddress.String() {
			for i := 0; i < 10; i++ {
				deleteErr := v.NftCollectionRepo.DeleteCollection(svcCtx, toAddress.String())
				if deleteErr == nil {
					break
				}
				log.Printf("‼️ Cant delete NFT Collection: %v, on error from DB, TRY: %v\n", toAddress.String(), i)
				time.Sleep(1 * time.Second)
			}
		}
		return &nftcollection.NftCollection{}, fmt.Errorf("error update user's balance due nft collection mint: %v", updErr)
	}

	msg := &tlb.ExternalMessage{
		DstAddr: marketAddress,
		Body:    msgData,
	}

	log.Println("Sending external message with hash:", hex.EncodeToString(msg.NormalizedHash()))

	msgErr := api.SendExternalMessage(apiCtx, msg)

	if msgErr != nil {
		// FYI: it can fail if not enough balance on contract
		if deployCfg.OwnerAddress == marketAddress.String() {
			for i := 0; i < 10; i++ {
				deleteErr := v.NftCollectionRepo.DeleteCollection(svcCtx, toAddress.String())
				if deleteErr == nil {
					break
				}
				log.Printf("‼️ Cant delete NFT Collection: %v, on error from DB, TRY: %v\n", toAddress.String(), i)
				time.Sleep(1 * time.Second)
			}
		}

		v.UserRepo.UpdateUserBalance(svcCtx, ownerAccount.UUID, ownerAccount.NanoTon-nanoTonsForDeploy+55000000) // if error return TON without FEES
		return &nftcollection.NftCollection{}, msgErr
	}

	log.Printf("NFT COLLECTION DEPLOYED AT ADDRESS: %v\n", toAddress.String())

	return nftCollection, nil
}
