package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	walletservice "github.com/rom6n/create-nft-go/internal/service/walletService"
)

type WalletHandler struct {
	WalletServiceRepo walletservice.WalletServiceRepository
}

func (v *WalletHandler) GetWalletData() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		walletAddress := c.Query("wallet-address")

		if walletAddress == "" {
			return c.Status(fiber.StatusBadRequest).SendString("Wallet Address is required")
		}

		foundWallet, svcErr := v.WalletServiceRepo.GetWalletByAddress(ctx, walletAddress)
		if svcErr != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Error while getting wallet data: %v", svcErr))
		}

		return c.Status(fiber.StatusOK).JSON(foundWallet)
	}
}

func (v *WalletHandler) RefreshWalletNftItems() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		walletAddress := c.Query("wallet-address")
		if walletAddress == "" {
			return c.Status(fiber.StatusBadRequest).SendString("Wallet Address is required")
		}

		nftItems, updateErr := v.WalletServiceRepo.UpdateWalletNftItems(ctx, walletAddress)
		if updateErr != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Error update NFT items: %v", updateErr))
		}

		return c.Status(fiber.StatusOK).JSON(nftItems)
	}
}
