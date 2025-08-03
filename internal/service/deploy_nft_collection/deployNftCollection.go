package deploynftcollection

import (
	"context"
	"crypto/ed25519"
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
	nftCollectionRepo          nftcollection.NftCollectionRepository
	userRepo                   user.UserRepository
	privateKey                 ed25519.PrivateKey
	liteClient                 *liteclient.ConnectionPool
	liteclientApi              ton.APIClientWrapped
	marketplaceContractAddress *address.Address
	nftCollectionContractCode  *cell.Cell
	nftItemContractCode        *cell.Cell
	timeout                    time.Duration
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
	Timeout                    time.Duration
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
		cfg.Timeout,
	}
}

func (v *deployNftCollectionServiceRepo) DeployNftCollection(ctx context.Context, deployCfg nftcollection.DeployCollectionCfg, ownerID int64) (*nftcollection.NftCollection, error) {
	svcCtx, cancel := context.WithTimeout(ctx, v.timeout)
	defer cancel()

	marketAddress := v.marketplaceContractAddress
	client := v.liteClient
	api := v.liteclientApi
	apiCtx := client.StickyContext(svcCtx)
	nanoTonForDeploy := uint64(50000000)
	nanoTonForFees := uint64(15000000)
	nilNftCollection := &nftcollection.NftCollection{}

	ownerAccount, userErr := v.userRepo.GetUserByID(svcCtx, ownerID)
	if userErr != nil {
		return nilNftCollection, userErr
	}

	if ownerAccount.NanoTon < (nanoTonForDeploy + nanoTonForFees) { // return if not enough TON
		return nilNftCollection, fmt.Errorf("not enough toncoins. Need %v nano ton more", (nanoTonForDeploy + nanoTonForFees - ownerAccount.NanoTon))
	}

	block, chainErr := api.CurrentMasterchainInfo(apiCtx)
	if chainErr != nil {
		return nilNftCollection, chainErr
	}

	response, methodErr := api.WaitForBlock(block.SeqNo).RunGetMethod(apiCtx, block, marketAddress, "seqno")
	if methodErr != nil {
		return nilNftCollection, methodErr
	}

	seqno := response.MustInt(0)

	content := nftcollectionutils.PackOffchainContentForNftCollection(deployCfg.CollectionContent, deployCfg.CommonContent)

	royaltyParams := nftcollectionutils.PackNftCollectionRoyaltyParams(deployCfg.RoyaltyDividend, deployCfg.RoyaltyDivisor, deployCfg.OwnerAddress)

	if deployCfg.OwnerAddress == nil {
		deployCfg.OwnerAddress = marketAddress
	}

	dataCell := nftcollectionutils.PackNftCollectionData(deployCfg.OwnerAddress, content, v.nftItemContractCode, royaltyParams)

	stateInit := generalcontractutils.PackStateInit(v.nftCollectionContractCode, dataCell)

	toAddress := generalcontractutils.CalculateAddress(0, stateInit)

	deployMsg := generalcontractutils.PackDeployMessage(toAddress, stateInit)

	validUntil := time.Now().Add(1 * time.Minute).Unix()

	msgData := marketutils.PackMessageToMarketplaceContract(v.privateKey, validUntil, seqno, 1, deployMsg)

	nftCollectionMetadata, metadataErr := nftcollectionutils.GetNftCollectionOffchainMetadata(deployCfg.CollectionContent)
	if metadataErr != nil {
		return nilNftCollection, metadataErr
	}

	nftCollection := nftcollection.New(toAddress.String(), ownerAccount.UUID, nftCollectionMetadata)

	// reducing the user's balance before deploy
	if updErr := v.userRepo.UpdateUserBalance(svcCtx, ownerAccount.UUID, ownerAccount.NanoTon-nanoTonForDeploy-nanoTonForFees); updErr != nil {
		return nilNftCollection, fmt.Errorf("error update user's balance before nft collection mint: %v", updErr)
	}

	msg := &tlb.ExternalMessage{
		DstAddr: marketAddress,
		Body:    msgData,
	}

	msgErr := api.SendExternalMessage(apiCtx, msg)

	if msgErr != nil {
		// FYI: it can fail if not enough balance on contract
		for i := 0; i < 10; i++ {
			if updErr := v.userRepo.UpdateUserBalance(svcCtx, ownerAccount.UUID, ownerAccount.NanoTon-uint64(nanoTonForFees)); updErr == nil {
				break
			}
			log.Printf("Error returning %v ton to user after nft collection deploy fail, try: %v\n", nanoTonForDeploy, i)
			if i == 9 {
				return nilNftCollection, fmt.Errorf("error returning ton to user & error sending deploy nft collection external message: %v", msgErr)
			}
			time.Sleep(1 * time.Second)
		}

		return nilNftCollection, fmt.Errorf("error sending deploy nft collection external message: %v", msgErr)
	}

	if deployCfg.OwnerAddress.Equals(v.marketplaceContractAddress) {
		for i := 0; i < 10; i++ {
			if createErr := v.nftCollectionRepo.CreateNftCollection(svcCtx, nftCollection); createErr == nil {
				break
			}
			log.Printf("Error adding nft collection to database, try: %v\n", i)
			if i == 9 {
				return nilNftCollection, fmt.Errorf("error adding nft collection to database")
			}
			time.Sleep(1 * time.Second)
		}
	}

	log.Printf("NFT COLLECTION DEPLOYED AT ADDRESS: %v\n", toAddress.String())

	return nftCollection, nil
}
