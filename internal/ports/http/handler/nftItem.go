package handler

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	nftitem "github.com/rom6n/create-nft-go/internal/domain/nft_item"
	mintnftitem "github.com/rom6n/create-nft-go/internal/service/mint_nft_item"
	withdrawnftitem "github.com/rom6n/create-nft-go/internal/service/withdraw_nft_item"
	"github.com/xssnick/tonutils-go/address"
)

type NftItemHandler struct {
	MintNftItemService     mintnftitem.MintNftItemServiceRepository
	WithdrawNftItemService withdrawnftitem.WithdrawNftItemServiceRepository
}

func (v *NftItemHandler) MintNftItem() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ownerWallet, content, fwdAmount, fwdMsg, nftCollectionAddress, ownerID, isTest := c.Query("owner-wallet"), c.Query("content"), c.Query("forward-amount"), c.Query("forward-message"), c.Query("nft-collection-address"), c.Query("owner-id"), c.Query("is-testnet")
		if content == "" || nftCollectionAddress == "" || ownerID == "" || isTest == "" {
			return c.Status(fiber.StatusBadRequest).SendString("content link, is testnet, owner id and nft collection address are required")
		}

		// ?owner-wallet=0QDU46qYz4rHAJhszrW9w6imF8p4Cw5dS1GpPTcJ9vqNSmnf&owner-id=5003727541&content=https://rom6n.github.io/mc1f/nft-c1-item-2.json&forward-amount=&forward-message=&nft-collection-address=EQBNQ_nUxOprp6Ak9FUo5HiM5XrW95u1y1QAL4659zi8rWVD&is-testnet=true

		var ownerAddress *address.Address
		if ownerWallet != "" {
			ownerAddress2, parseAddrErr := address.ParseAddr(ownerWallet)
			if parseAddrErr != nil {
				return c.Status(fiber.StatusBadRequest).SendString("owner wallet is not valid address")
			} else {
				ownerAddress = ownerAddress2
			}
		}

		isTestnet, parseBoolErr := strconv.ParseBool(isTest)
		if parseBoolErr != nil {
			c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Error parse is-testnet to bool: %v", parseBoolErr))
		}

		nftCollectionAddr, parseAddrErr := address.ParseAddr(nftCollectionAddress)
		if parseAddrErr != nil {
			return c.Status(fiber.StatusBadRequest).SendString("nft collection is not valid address")
		}

		ownerIDInt64, parseErr := strconv.ParseInt(ownerID, 0, 64)
		if parseErr != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("error parsing user id to int64: %v", parseErr))
		}

		var forvardAmount uint64
		if fwdAmount != "" {
			if forvardAmountParsed, parseErr := strconv.ParseUint(fwdAmount, 0, 64); parseErr != nil {
				return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("error parsing forward amount to uint64: %v", parseErr))
			} else {
				forvardAmount = forvardAmountParsed
			}
		}

		mintCfg := nftitem.MintNftItemCfg{
			OwnerAddress:   ownerAddress,
			Content:        content,
			ForwardAmount:  forvardAmount,
			ForwardMessage: fwdMsg,
		}

		nftItem, mintErr := v.MintNftItemService.MintNftItem(c.Context(), nftCollectionAddr, mintCfg, ownerIDInt64, isTestnet)
		if mintErr != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("error minting nft item: %v", mintErr))
		}

		return c.Status(fiber.StatusOK).JSON(nftItem)
	}
}

func (v *NftItemHandler) WithdrawNftItem() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		nftItemAddressStr, WithdrawToAddressStr, ownerIDStr, isTest := c.Params("address"), c.Query("withdraw-to"), c.Query("owner-id"), c.Query("is-testnet")
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

		nftItemAddress, parseAddrErr := address.ParseAddr(nftItemAddressStr)
		if parseAddrErr != nil {
			return c.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("nft item is not valid address: %v\n%v", nftItemAddressStr, parseAddrErr))
		}

		newOwnerAddress, parseAddrErr := address.ParseAddr(WithdrawToAddressStr)
		if parseAddrErr != nil {
			return c.Status(fiber.StatusBadRequest).SendString("new owner is not valid address")
		}

		if withdrawErr := v.WithdrawNftItemService.WithdrawNftItem(ctx, nftItemAddress, newOwnerAddress, int64(ownerID), isTestnet); withdrawErr != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("error withdrawing: %v", withdrawErr))
		}

		return c.Status(fiber.StatusOK).SendString("Successfully withdrawed nft item")
	}
}
