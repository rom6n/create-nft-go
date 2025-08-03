package handler

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	nftcollection "github.com/rom6n/create-nft-go/internal/domain/nft_collection"
	deploynftcollection "github.com/rom6n/create-nft-go/internal/service/deploy_nft_collection"
	nftcollectionservice "github.com/rom6n/create-nft-go/internal/service/nft_collection_service"
	"github.com/xssnick/tonutils-go/address"
)

type NftCollectionHandler struct {
	NftCollectionService       nftcollectionservice.NftCollectionServiceRepository
	DeployNftCollectionService deploynftcollection.DeployNftCollectionServiceRepository
}

func (v *NftCollectionHandler) DeployNftCollection() fiber.Handler { 
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		ownerWallet, ownerIDStr, commonContent, collectionContent, royaltyDividendStr, royaltyDivisorStr :=
			c.Query("owner-wallet"), c.Query("owner-id"), c.Query("common-content"), c.Query("collection-content"), c.Query("royalty-dividend"), c.Query("royalty-divisor")

		if ownerIDStr == "" || commonContent == "" || collectionContent == "" || royaltyDividendStr == "" || royaltyDivisorStr == "" {
			return c.Status(fiber.StatusBadRequest).SendString("ownerID, common content, collection content, royalty dividend, royalty divisor are required")
		}

		var ownerAddress *address.Address
		if ownerWallet != "" {
			if ownerAddress2, parseAddrErr := address.ParseAddr(ownerWallet); parseAddrErr != nil {
				return c.Status(fiber.StatusBadRequest).SendString("owner wallet is not valid address")
			} else {
				ownerAddress = ownerAddress2
			}
		}

		// ?owner-wallet=0QDU46qYz4rHAJhszrW9w6imF8p4Cw5dS1GpPTcJ9vqNSmnf&owner-id=5003727541&common-content=https://&collection-content=https://rom6n.github.io/mc1f/nft-c1-collection.json&royalty-dividend=20&royalty-divisor=100

		royaltyDividend, parseErr := strconv.ParseUint(royaltyDividendStr, 0, 16)
		royaltyDivisor, parseErr2 := strconv.ParseUint(royaltyDivisorStr, 0, 16)
		if parseErr != nil || parseErr2 != nil {
			c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Error parse to uint16: %v. Error 2: %v", parseErr, parseErr2))
		}

		ownerID, parseErr := strconv.Atoi(ownerIDStr)
		if parseErr != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Error parse to int: %v", parseErr))
		}

		collectionCfg := nftcollection.DeployCollectionCfg{
			OwnerAddress:      ownerAddress,
			CommonContent:     commonContent,
			CollectionContent: collectionContent,
			RoyaltyDividend:   uint16(royaltyDividend),
			RoyaltyDivisor:    uint16(royaltyDivisor),
		}

		if v.DeployNftCollectionService == nil || v.NftCollectionService == nil {
			return c.Status(fiber.StatusInternalServerError).SendString("DeployNftCollectionService or NftCollectionService is not initialized")
		}

		collection, deployErr := v.DeployNftCollectionService.DeployNftCollection(ctx, collectionCfg, int64(ownerID))
		if deployErr != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Error while deploying nft collection: %v", deployErr))
		}

		return c.Status(fiber.StatusOK).JSON(collection)
	}
}

func (v *NftCollectionHandler) DeployMarketContract() fiber.Handler { // Пока что деньги будут списываться с подключенного кошелька. Потом добавлю баланс в приложении
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		if err := v.NftCollectionService.DeployMarketplaceContract(ctx); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Error deploy market: %v", err))
		}

		return c.Status(fiber.StatusOK).SendString("Successfully deployed")
	}
}
