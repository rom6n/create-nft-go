package handler

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	nftcollection "github.com/rom6n/create-nft-go/internal/domain/nft_collection"
	deploynftcollection "github.com/rom6n/create-nft-go/internal/service/deploy_nft_collection"
	nftcollectionservice "github.com/rom6n/create-nft-go/internal/service/nft_collection_service"
	withdrawnftcollection "github.com/rom6n/create-nft-go/internal/service/withdraw_nft_collection"
	"github.com/xssnick/tonutils-go/address"
)

type NftCollectionHandler struct {
	NftCollectionService         nftcollectionservice.NftCollectionServiceRepository
	DeployNftCollectionService   deploynftcollection.DeployNftCollectionServiceRepository
	WithdrawNftCollectionService withdrawnftcollection.WithdrawNftCollectionServiceRepository
}

func (v *NftCollectionHandler) DeployNftCollection() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		ownerWallet, ownerIDStr, collectionContent, royaltyDividendStr, royaltyDivisorStr, isTest :=
			c.Query("owner-wallet"), c.Query("owner-id"), c.Query("collection-content"), c.Query("royalty-dividend"), c.Query("royalty-divisor"), c.Query("is-testnet")

		if ownerIDStr == "" || collectionContent == "" || royaltyDividendStr == "" || royaltyDivisorStr == "" || isTest == "" {
			return c.Status(fiber.StatusBadRequest).SendString("owner id, is testnet, collection content, royalty dividend, royalty divisor are required")
		}

		var ownerAddress *address.Address
		if ownerWallet != "" {
			if ownerAddress2, parseAddrErr := address.ParseAddr(ownerWallet); parseAddrErr != nil {
				return c.Status(fiber.StatusBadRequest).SendString("owner wallet is not valid address")
			} else {
				ownerAddress = ownerAddress2
			}
		}

		// ?owner-wallet=0QDU46qYz4rHAJhszrW9w6imF8p4Cw5dS1GpPTcJ9vqNSmnf&owner-id=5003727541&common-content=https://&collection-content=https://rom6n.github.io/mc1f/nft-c1-collection.json&royalty-dividend=20&royalty-divisor=100&is-testnet=true
		isTestnet, parseBoolErr := strconv.ParseBool(isTest)
		if parseBoolErr != nil {
			c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Error parse is-testnet to bool: %v", parseBoolErr))
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

		commonContent := "https://" // common content will always start with https://
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

		collection, deployErr := v.DeployNftCollectionService.DeployNftCollection(ctx, collectionCfg, int64(ownerID), isTestnet)
		if deployErr != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Error while deploying nft collection: %v", deployErr))
		}

		return c.Status(fiber.StatusOK).JSON(collection)
	}
}

func (v *NftCollectionHandler) WithdrawNftCollection() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		collectionAddressStr, WithdrawToAddressStr, ownerIDStr, isTest := c.Params("address"), c.Query("withdraw-to"), c.Query("owner-id"), c.Query("is-testnet")
		if WithdrawToAddressStr == "" || isTest == "" || ownerIDStr == "" {
			return c.Status(fiber.StatusBadRequest).SendString("withdraw to, owner id and is testnet are required")
		}

		ownerID, parseErr := strconv.Atoi(ownerIDStr)
		if parseErr != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Error parse to int: %v", parseErr))
		}

		isTestnet, parseBoolErr := strconv.ParseBool(isTest)
		if parseBoolErr != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Error parse is-testnet to bool: %v", parseBoolErr))
		}

		nftCollectionAddress, parseAddrErr := address.ParseAddr(collectionAddressStr)
		if parseAddrErr != nil {
			return c.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("nft collection is not valid address: %v\n%v", collectionAddressStr, parseAddrErr))
		}

		newOwnerAddress, parseAddrErr := address.ParseAddr(WithdrawToAddressStr)
		if parseAddrErr != nil {
			return c.Status(fiber.StatusBadRequest).SendString("new owner is not valid address")
		}

		if withdrawErr := v.WithdrawNftCollectionService.WithdrawNftCollection(ctx, nftCollectionAddress, newOwnerAddress, int64(ownerID), isTestnet); withdrawErr != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("error withdrawing: %v", withdrawErr))
		}

		return c.Status(fiber.StatusOK).SendString("Successfully withdrawed nft collection")
	}
}
