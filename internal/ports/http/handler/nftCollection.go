package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	nftcollection "github.com/rom6n/create-nft-go/internal/domain/nftCollection"
	nftcollectionservice "github.com/rom6n/create-nft-go/internal/service/nftCollectionService"
)

type NftCollectionHandler struct {
	NftCollectionService nftcollectionservice.NftCollectionServiceRepository
}

func (v *NftCollectionHandler) mintNftCollection() fiber.Handler { // Пока что деньги будут списываться с подключенного кошелька. Потом добавлю баланс в приложении
	return func(c *fiber.Ctx) error {
		var collectionCfg nftcollection.MintCollectionCfg
		parseErr := c.BodyParser(&collectionCfg)
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("Error. NFT Collection config must match cfg struct: %v", parseErr))
		}

		return c.Status(fiber.StatusOK).SendString("Successfully minted")
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
