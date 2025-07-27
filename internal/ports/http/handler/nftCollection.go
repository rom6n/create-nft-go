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

func (v *NftCollectionHandler) DeployNftCollection() fiber.Handler { // Пока что деньги будут списываться с подключенного кошелька. Потом добавлю баланс в приложении
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		//var collectionCfg nftcollection.DeployCollectionCfg
		//parseErr := c.BodyParser(&collectionCfg)
		//if parseErr != nil {
		//return c.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("Error. NFT Collection config must match cfg struct: %v", parseErr))
		//}

		//FOR TEST ONLY
		collectionCfg := nftcollection.DeployCollectionCfg{
			Owner:             "0QDU46qYz4rHAJhszrW9w6imF8p4Cw5dS1GpPTcJ9vqNSmnf",
			CommonContent:     "https://t.me/MrRoman",
			CollectionContent: "https://t.me/MrRoman",
			RoyaltyDividend:   2,
			RoyaltyDivisor:    100,
		}
		collection, deployErr := v.NftCollectionService.DeployNftCollection(ctx, collectionCfg, 5003727541)
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
