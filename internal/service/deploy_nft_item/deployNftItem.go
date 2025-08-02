package deploynftitem

import (
	"context"
	"crypto/ed25519"
	"fmt"
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
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type deployNftItemServiceRepository interface {
	//DeployNftItem(ctx context.Context) (*nft.NftItem, error)
}

type deployNftItemServiceRepo struct {
	nftCollectionRepo          nftcollection.NftCollectionRepository
	nftItemRepo                nft.NftItemRepository
	userRepo                   user.UserRepository
	nftItemCode                *cell.Cell
	liteClient                 *liteclient.ConnectionPool
	liteclientApi              ton.APIClientWrapped
	marketplaceContractAddress *address.Address
	privateKey                 ed25519.PrivateKey
	timeout                    time.Duration
}

type DeployNftItemServiceCfg struct {
	NftCollectionRepo          nftcollection.NftCollectionRepository
	NftItemRepo                nft.NftItemRepository
	UserRepo                   user.UserRepository
	NftItemCode                *cell.Cell
	LiteClient                 *liteclient.ConnectionPool
	LiteclientApi              ton.APIClientWrapped
	MarketplaceContractAddress *address.Address
	PrivateKey                 ed25519.PrivateKey
	Timeout                    time.Duration
}

func New(cfg DeployNftItemServiceCfg) deployNftItemServiceRepository {
	return deployNftItemServiceRepo{
		nftCollectionRepo:          cfg.NftCollectionRepo,
		nftItemRepo:                cfg.NftItemRepo,
		userRepo:                   cfg.UserRepo,
		nftItemCode:                cfg.NftItemCode,
		liteClient:                 cfg.LiteClient,
		liteclientApi:              cfg.LiteclientApi,
		marketplaceContractAddress: cfg.MarketplaceContractAddress,
		privateKey:                 cfg.PrivateKey,
		timeout:                    cfg.Timeout,
	}
}

func (v *deployNftItemServiceRepo) GetContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, v.timeout)
}

func (v *deployNftItemServiceRepo) DeployNftItem(ctx context.Context, nftCollectionAddress *address.Address, cfg nft.DeployNftItemCfg, ownerID int64) (*nft.NftItem, error) {
	svcCtx, cancel := v.GetContext(ctx)
	defer cancel()

	nilNftItem := &nft.NftItem{}

	client := v.liteClient
	api := v.liteclientApi

	apiCtx := client.StickyContext(svcCtx)

	ownerAccount, userErr := v.userRepo.GetUserByID(svcCtx, ownerID)
	if userErr != nil {
		return nilNftItem, userErr
	}

	if cfg.OwnerAddress.String() == "" {
		cfg.OwnerAddress = v.marketplaceContractAddress
	}

	block, bErr := api.GetMasterchainInfo(apiCtx)
	if bErr != nil {
		return nilNftItem, fmt.Errorf("error getting masterchain info: %v", bErr)
	}

	response, responseErr := api.WaitForBlock(block.SeqNo).RunGetMethod(apiCtx, block, v.marketplaceContractAddress, "seqno")
	if responseErr != nil {
		return nilNftItem, fmt.Errorf("error getting marketplace contract seqno: %v", responseErr)
	}

	response2, response2Err := api.WaitForBlock(block.SeqNo).RunGetMethod(apiCtx, block, nftCollectionAddress, "get_collection_data")
	if response2Err != nil {
		return nilNftItem, fmt.Errorf("error getting marketplace contract seqno: %v", response2Err)
	}

	seqno := response.MustInt(0)
	nextItemIndex := response2.MustInt(0)
	nftCollectionMetadataLink := response2.MustCell(1).BeginParse().MustLoadStringSnake()
	nftCollectionOwnerSlice := response2.MustSlice(2)

	nftCollectionOwnerAddress := address.MustParseRawAddr(nftCollectionOwnerSlice.String())

	if !nftCollectionOwnerAddress.Equals(v.marketplaceContractAddress) {
		return nilNftItem, fmt.Errorf("error, to deploy nft item marketplace contract must be nft collection owner")
	}

	// ToDo
	// Сделать проверку наличия в БД
	//
	// Сделать проверку баланса пользователя
	//

	nftCollectionMetadata, metaErr := nftcollectionutils.GetNftCollectionMetadataByLink(nftCollectionMetadataLink)
	if metaErr != nil {
		return nilNftItem, metaErr
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
	nftItemMetadata, metaErr := nftitemutils.GetNftItemMetadataByLin(cfg.Content)
	if metaErr != nil {
		return nilNftItem, fmt.Errorf("error parse nft item metadata: %v", metaErr)
	}

	nftItem := nft.NftItem{
		Address:           nftItemAddress.String(),
		Index:             nextItemIndex.Int64(),
		CollectionAddress: nftCollectionAddress.String(),
		CollectionName:    nftCollectionMetadata.Name,
		Owner:             ownerAccount.UUID,
		Metadata:          *nftItemMetadata,
	}

	// ToDo
	// Добавить Nft Item в БД
	//
	// Списать средства пользователя
	//
	//

	msg := &tlb.ExternalMessage{
		DstAddr: v.marketplaceContractAddress,
		Body:    marketplaceContractMsg,
	}

	msgErr := api.SendExternalMessage(apiCtx, msg)
	if msgErr != nil {
		// FIY: not enough balance on contract
		// ToDo
		//
		// При ошибке удалить nft item из БД
		//
		// Вернуть средства пользователю
		//

		return nilNftItem, fmt.Errorf("error sending deploy nft item external message: %v", msgErr)
	}

	return &nftItem, nil
}
