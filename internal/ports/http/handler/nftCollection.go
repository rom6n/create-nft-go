package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	nftcollection "github.com/rom6n/create-nft-go/internal/domain/nftCollection"
	nftcollectionservice "github.com/rom6n/create-nft-go/internal/service/nftCollectionService"
)

type NftCollectionHandler struct {
	nftCollectionService nftcollectionservice.NftCollectionServiceRepository
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
