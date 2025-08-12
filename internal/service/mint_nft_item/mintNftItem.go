package mintnftitem

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"log"
	"time"

	nftcollection "github.com/rom6n/create-nft-go/internal/domain/nft_collection"
	nft "github.com/rom6n/create-nft-go/internal/domain/nft_item"
	"github.com/rom6n/create-nft-go/internal/domain/user"
	generalcontractutils "github.com/rom6n/create-nft-go/internal/utils/contract_utils/general_contract_utils"
	marketutils "github.com/rom6n/create-nft-go/internal/utils/contract_utils/market_utils"
	nftcollectionutils "github.com/rom6n/create-nft-go/internal/utils/contract_utils/nft_collection_utils"
	nftitemutils "github.com/rom6n/create-nft-go/internal/utils/contract_utils/nft_item_utils"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	tonnft "github.com/xssnick/tonutils-go/ton/nft"
	"github.com/xssnick/tonutils-go/tvm/cell"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type MintNftItemServiceRepository interface {
	MintNftItem(ctx context.Context, nftCollectionAddress *address.Address, cfg nft.MintNftItemCfg, ownerID int64, isTestnet bool) (*nft.NftItem, error)
}

type mintNftItemServiceRepo struct {
	nftCollectionRepo                 nftcollection.NftCollectionRepository
	nftItemRepo                       nft.NftItemRepository
	userRepo                          user.UserRepository
	nftItemCode                       *cell.Cell
	testnetLiteClient                 *liteclient.ConnectionPool
	mainnetLiteClient                 *liteclient.ConnectionPool
	testnetLiteApi                    ton.APIClientWrapped
	mainnetLiteApi                    ton.APIClientWrapped
	testnetMarketplaceContractAddress *address.Address
	mainnetMarketplaceContractAddress *address.Address
	privateKey                        ed25519.PrivateKey
	timeout                           time.Duration
}

type MintNftItemServiceCfg struct {
	NftCollectionRepo                 nftcollection.NftCollectionRepository
	NftItemRepo                       nft.NftItemRepository
	UserRepo                          user.UserRepository
	NftItemCode                       *cell.Cell
	TestnetLiteClient                 *liteclient.ConnectionPool
	MainnetLiteClient                 *liteclient.ConnectionPool
	TestnetLiteApi                    ton.APIClientWrapped
	MainnetLiteApi                    ton.APIClientWrapped
	TestnetMarketplaceContractAddress *address.Address
	MainnetMarketplaceContractAddress *address.Address
	PrivateKey                        ed25519.PrivateKey
	Timeout                           time.Duration
}

func New(cfg MintNftItemServiceCfg) MintNftItemServiceRepository {
	return &mintNftItemServiceRepo{
		nftCollectionRepo:                 cfg.NftCollectionRepo,
		nftItemRepo:                       cfg.NftItemRepo,
		userRepo:                          cfg.UserRepo,
		nftItemCode:                       cfg.NftItemCode,
		testnetLiteClient:                 cfg.TestnetLiteClient,
		mainnetLiteClient:                 cfg.MainnetLiteClient,
		testnetLiteApi:                    cfg.TestnetLiteApi,
		mainnetLiteApi:                    cfg.MainnetLiteApi,
		testnetMarketplaceContractAddress: cfg.TestnetMarketplaceContractAddress,
		mainnetMarketplaceContractAddress: cfg.MainnetMarketplaceContractAddress,
		privateKey:                        cfg.PrivateKey,
		timeout:                           cfg.Timeout,
	}
}

func (v *mintNftItemServiceRepo) getContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, v.timeout)
}

func (v *mintNftItemServiceRepo) MintNftItem(ctx context.Context, nftCollectionAddress *address.Address, cfg nft.MintNftItemCfg, ownerID int64, isTestnet bool) (*nft.NftItem, error) {
	svcCtx, cancel := v.getContext(ctx)
	defer cancel()

	// сделать проверку на fwd amount
	nanoTonForMint := uint64(60000000)
	nanoTonForFees := uint64(5000000)
	if cfg.ForwardAmount >= 1 {
		nanoTonForMint += cfg.ForwardAmount
	}

	client := v.testnetLiteClient
	api := v.testnetLiteApi
	marketplaceContractAddress := v.testnetMarketplaceContractAddress

	nftCollectionAddress.SetTestnetOnly(isTestnet)

	// checking for nft collection in DB
	if _, getErr := v.nftCollectionRepo.GetNftCollectionByAddress(svcCtx, nftCollectionAddress.String()); getErr != nil {
		if getErr == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("nft collection isnt in database")
		}
		return nil, fmt.Errorf("find error in database: %v", getErr)
	}

	if !isTestnet {
		client = v.mainnetLiteClient
		api = v.mainnetLiteApi
		marketplaceContractAddress = v.mainnetMarketplaceContractAddress
	}

	apiCtx := client.StickyContext(svcCtx)
	nftCollectionAddress.SetTestnetOnly(isTestnet)

	ownerAccount, userErr := v.userRepo.GetUserByID(svcCtx, ownerID)
	if userErr != nil {
		return nil, userErr
	}

	// checking for user have enough ton
	if ownerAccount.NanoTon < (nanoTonForMint + nanoTonForFees) {
		return nil, fmt.Errorf("error not enough ton on user's balance. need %v more", nanoTonForMint-ownerAccount.NanoTon)
	}

	if cfg.OwnerAddress == nil {
		cfg.OwnerAddress = marketplaceContractAddress
	}

	block, bErr := api.GetMasterchainInfo(apiCtx)
	if bErr != nil {
		return nil, fmt.Errorf("error getting masterchain info: %v", bErr)
	}

	response, responseErr := api.WaitForBlock(block.SeqNo).RunGetMethod(apiCtx, block, marketplaceContractAddress, "seqno")
	if responseErr != nil {
		return nil, fmt.Errorf("error getting marketplace contract seqno: %v", responseErr)
	}

	collectionClient := tonnft.NewCollectionClient(api, nftCollectionAddress)
	collectionData, dataErr := collectionClient.GetCollectionDataAtBlock(apiCtx, block)
	if dataErr != nil {
		return nil, fmt.Errorf("fail getting nft collection data method: %v", dataErr)
	}

	seqno, seqnoErr := response.Int(0)
	if seqnoErr != nil {
		return nil, fmt.Errorf("marketplace method seqno returned not a seqno: %v", seqnoErr)
	}

	nextItemIndex := collectionData.NextItemIndex
	collectionContent, cellErr := collectionData.Content.ContentCell()
	if cellErr != nil {
		return nil, fmt.Errorf("nft collection data method not returned content cell: %v", cellErr)
	}

	// ✨ ToDo: В будущем можно будет добавить поддержку onchain метадаты ✨
	collectionContentSlice := collectionContent.BeginParse()
	if skipOffchainTag, tagErr := collectionContentSlice.LoadUInt(8); skipOffchainTag != 1 || tagErr != nil {
		return nil, fmt.Errorf("want offchain metadata tag, have other: %v", tagErr)
	}
	nftCollectionMetadataLink, offChainMetadataErr := collectionContentSlice.LoadStringSnake()

	if offChainMetadataErr != nil {
		return nil, fmt.Errorf("want string, have other: %v", offChainMetadataErr)
	}

	nftCollectionOwnerAddress := collectionData.OwnerAddress

	// checking for market contract is nft collection owner
	if !nftCollectionOwnerAddress.Equals(marketplaceContractAddress) {
		return nil, fmt.Errorf("marketplace contract must be a nft collection owner to deploy nft item")
	}

	nftCollectionMetadata, metaErr := nftcollectionutils.GetNftCollectionOffchainMetadata(nftCollectionMetadataLink)
	if metaErr != nil {
		return nil, metaErr
	}

	nftItemMetadata, metaErr := nftitemutils.GetNftItemOffchainMetadata(cfg.Content)
	if metaErr != nil {
		return nil, fmt.Errorf("error parse nft item metadata: %v", metaErr)
	}

	deployNftItemMsg := nftcollectionutils.PackDeployNftItemMessage(nftCollectionAddress, nextItemIndex.Uint64(), cfg)

	marketplaceContractMsg := marketutils.PackMessageToMarketplaceContract(v.privateKey, time.Now().Add(1*time.Minute).Unix(), seqno, 1, deployNftItemMsg)

	stateInit := generalcontractutils.PackStateInit(v.nftItemCode,
		cell.BeginCell().
			MustStoreUInt(nextItemIndex.Uint64(), 64).
			MustStoreAddr(nftCollectionAddress).
			EndCell(),
	)

	nftItemAddress := generalcontractutils.CalculateAddress(0, stateInit)
	nftItemAddress.SetTestnetOnly(isTestnet)

	nftItem := nft.New(
		nftItemAddress.String(),
		nextItemIndex.Int64(),
		nftCollectionAddress.String(),
		nftCollectionMetadata.Name,
		ownerAccount.UUID,
		nftItemMetadata,
		isTestnet,
	)
	
	// ДОБАВИТЬ ТРАНЗАКЦИЮ
	// reducing the user's balance
	if updErr := v.userRepo.UpdateUserBalance(svcCtx, ownerAccount.UUID, ownerAccount.NanoTon-nanoTonForMint-nanoTonForFees); updErr != nil {
		return nil, fmt.Errorf("error reducing user's balance for nft item mint: %v", updErr)
	}

	msg := &tlb.ExternalMessage{
		DstAddr: marketplaceContractAddress,
		Body:    marketplaceContractMsg,
	}

	msgErr := api.SendExternalMessage(apiCtx, msg)
	if msgErr != nil {
		// FIY: not enough balance on contract
		for i := 0; i < 10; i++ {
			if updErr := v.userRepo.UpdateUserBalance(svcCtx, ownerAccount.UUID, ownerAccount.NanoTon-nanoTonForFees); updErr == nil {
				break
			}
			log.Printf("Error returning %v ton to user, try: %v\n", nanoTonForMint, i)
			if i == 9 {
				return nil, fmt.Errorf("error returning ton to user & error sending mint nft item external message: %v", msgErr)
			}
			time.Sleep(1 * time.Second)
		}

		return nil, fmt.Errorf("error sending mint nft item by external message: %v", msgErr)
	}

	if cfg.OwnerAddress.Equals(marketplaceContractAddress) {
		for i := 0; i < 10; i++ {
			if createErr := v.nftItemRepo.CreateNftItem(svcCtx, nftItem); createErr == nil {
				break
			}
			log.Printf("Error adding nft item to database, try: %v\n", i)
			if i == 9 {
				return nil, fmt.Errorf("error adding nft item to database")
			}
			time.Sleep(1 * time.Second)
		}
	}

	return nftItem, nil
}
