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
	nftcollectionutils "github.com/rom6n/create-nft-go/internal/utils/contract_utils/nft_collection_utils"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type DeployNftCollectionServiceRepository interface {
	DeployNftCollection(ctx context.Context, deployCfg nftcollection.DeployCollectionCfg, ownerID int64, isTestnet bool) (*nftcollection.NftCollection, error)
}

type deployNftCollectionServiceRepo struct {
	nftCollectionRepo         nftcollection.NftCollectionRepository
	userRepo                  user.UserRepository
	privateKey                ed25519.PrivateKey
	testnetLiteClient         *liteclient.ConnectionPool
	mainnetLiteClient         *liteclient.ConnectionPool
	testnetLiteApi            ton.APIClientWrapped
	mainnetLiteApi            ton.APIClientWrapped
	testnetWallet             *wallet.Wallet
	mainnetWallet             *wallet.Wallet
	nftCollectionContractCode *cell.Cell
	nftItemContractCode       *cell.Cell
	timeout                   time.Duration
}

type DeployNftCollectionServiceCfg struct {
	NftCollectionRepo         nftcollection.NftCollectionRepository
	UserRepo                  user.UserRepository
	PrivateKey                ed25519.PrivateKey
	TestnetLiteClient         *liteclient.ConnectionPool
	MainnetLiteClient         *liteclient.ConnectionPool
	MainnetLiteApi            ton.APIClientWrapped
	TestnetLiteApi            ton.APIClientWrapped
	TestnetWallet             *wallet.Wallet
	MainnetWallet             *wallet.Wallet
	NftCollectionContractCode *cell.Cell
	NftItemContractCode       *cell.Cell
	Timeout                   time.Duration
}

func New(cfg DeployNftCollectionServiceCfg) DeployNftCollectionServiceRepository {
	return &deployNftCollectionServiceRepo{
		cfg.NftCollectionRepo,
		cfg.UserRepo,
		cfg.PrivateKey,
		cfg.TestnetLiteClient,
		cfg.MainnetLiteClient,
		cfg.TestnetLiteApi,
		cfg.MainnetLiteApi,
		cfg.TestnetWallet,
		cfg.MainnetWallet,
		cfg.NftCollectionContractCode,
		cfg.NftItemContractCode,
		cfg.Timeout,
	}
}

func (v *deployNftCollectionServiceRepo) DeployNftCollection(ctx context.Context, deployCfg nftcollection.DeployCollectionCfg, ownerID int64, isTestnet bool) (*nftcollection.NftCollection, error) {
	svcCtx, cancel := context.WithTimeout(ctx, v.timeout)
	defer cancel()

	walletAddress := v.testnetWallet.WalletAddress()
	client := v.testnetLiteClient
	w := v.testnetWallet

	if !isTestnet {
		walletAddress = v.mainnetWallet.WalletAddress()
		client = v.mainnetLiteClient
		w = v.mainnetWallet
	}

	apiCtx := client.StickyContext(svcCtx)
	nanoTonForDeploy := uint64(50000000)
	nanoTonForFees := uint64(15000000)

	ownerAccount, userErr := v.userRepo.GetUserByID(svcCtx, ownerID)
	if userErr != nil {
		return nil, userErr
	}

	if ownerAccount.NanoTon < (nanoTonForDeploy + nanoTonForFees) { // return if not enough TON
		return nil, fmt.Errorf("not enough toncoins. Need %v nano ton more", (nanoTonForDeploy + nanoTonForFees - ownerAccount.NanoTon))
	}

	content := nftcollectionutils.PackOffchainContentForNftCollection(deployCfg.CollectionContent, deployCfg.CommonContent)

	royaltyParams := nftcollectionutils.PackNftCollectionRoyaltyParams(deployCfg.RoyaltyDividend, deployCfg.RoyaltyDivisor, deployCfg.OwnerAddress)

	if deployCfg.OwnerAddress == nil {
		deployCfg.OwnerAddress = walletAddress
	}

	dataCell := nftcollectionutils.PackNftCollectionData(deployCfg.OwnerAddress, content, v.nftItemContractCode, royaltyParams)

	stateInit := generalcontractutils.PackStateInit(v.nftCollectionContractCode, dataCell)

	toAddress := generalcontractutils.CalculateAddress(0, stateInit)
	toAddress.SetTestnetOnly(isTestnet)

	deployMsg := generalcontractutils.PackDeployMessage(toAddress, stateInit)

	nftCollectionMetadata, metadataErr := nftcollectionutils.GetNftCollectionOffchainMetadata(deployCfg.CollectionContent)
	if metadataErr != nil {
		return nil, metadataErr
	}

	nftCollection := nftcollection.New(toAddress.String(), ownerAccount.UUID, nftCollectionMetadata, isTestnet)

	// reducing the user's balance before deploy
	if updErr := v.userRepo.UpdateUserBalance(svcCtx, ownerAccount.UUID, ownerAccount.NanoTon-nanoTonForDeploy-nanoTonForFees); updErr != nil {
		return nil, fmt.Errorf("error update user's balance before nft collection mint: %v", updErr)
	}

	msg := wallet.Message{
		Mode:            1,
		InternalMessage: deployMsg,
	}

	msgErr := w.Send(apiCtx, &msg, true)

	if msgErr != nil {
		// FYI: it can fail if not enough balance on contract
		for i := 0; i < 10; i++ {
			if updErr := v.userRepo.UpdateUserBalance(svcCtx, ownerAccount.UUID, ownerAccount.NanoTon-uint64(nanoTonForFees)); updErr == nil {
				break
			}
			log.Printf("Error returning %v ton to user after nft collection deploy fail, try: %v\n", nanoTonForDeploy, i)
			if i == 9 {
				return nil, fmt.Errorf("error returning ton to user & error sending deploy nft collection external message: %v", msgErr)
			}
			time.Sleep(1 * time.Second)
		}

		return nil, fmt.Errorf("error sending deploy nft collection external message: %v", msgErr)
	}

	if deployCfg.OwnerAddress.Equals(walletAddress) {
		for i := 0; i < 10; i++ {
			if createErr := v.nftCollectionRepo.CreateNftCollection(svcCtx, nftCollection); createErr == nil {
				break
			}
			log.Printf("Error adding nft collection to database, try: %v\n", i)
			if i == 9 {
				return nil, fmt.Errorf("error adding nft collection to database")
			}
			time.Sleep(1 * time.Second)
		}
	}

	log.Printf("Is Testnet: %v. NFT COLLECTION, DEPLOYED AT ADDRESS: %v\n", isTestnet, toAddress.String())

	return nftCollection, nil
}
