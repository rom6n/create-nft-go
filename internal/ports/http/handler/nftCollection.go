package handler

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	nftcollection "github.com/rom6n/create-nft-go/internal/domain/nftCollection"
	nftcollectionservice "github.com/rom6n/create-nft-go/internal/service/nftCollectionService"
)

type NftCollectionHandler struct {
	NftCollectionService nftcollectionservice.NftCollectionServiceRepository
}

func (v *NftCollectionHandler) DeployNftCollection() fiber.Handler { // Пока что деньги будут списываться с подключенного кошелька. Потом добавлю баланс в приложении
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		ownerWallet, ownerIDStr, commonContent, collectionContent, royaltyDividendStr, royaltyDivisorStr :=
			c.Params("owner"), c.Params("owner-id"), c.Params("common-content"), c.Params("collection-content"), c.Params("royalty-dividend"), c.Params("royalty-divisor")

		if ownerIDStr == "" || commonContent == "" || collectionContent == "" || royaltyDividendStr == "" || royaltyDivisorStr == "" {
			c.Status(fiber.StatusBadRequest).SendString("ownerID, common content, collection content, royalty dividend, royalty divisor are required")
		}

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
			Owner:             ownerWallet,
			CommonContent:     commonContent,
			CollectionContent: collectionContent,
			RoyaltyDividend:   uint16(royaltyDividend),
			RoyaltyDivisor:    uint16(royaltyDivisor),
		}

		collection, deployErr := v.NftCollectionService.DeployNftCollection(ctx, collectionCfg, int64(ownerID))
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
