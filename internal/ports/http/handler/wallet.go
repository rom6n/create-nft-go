package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/rom6n/create-nft-go/internal/domain/wallet"
	"github.com/rom6n/create-nft-go/internal/ports/http/api/ton"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type WalletHandler struct {
	WalletDB wallet.WalletRepository
	TonApi   ton.TonApiRepository
}

func (h *WalletHandler) GetWalletData() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		walletAddress := c.Query("wallet-address")

		if walletAddress == "" {
			return c.Status(fiber.StatusBadRequest).SendString("Wallet Address is required")
		}

		foundWallet, dbErr := h.WalletDB.GetWalletByAddress(ctx, walletAddress)
		if dbErr != nil {
			// добавляем кошелек в БД если его нет
			if dbErr == mongo.ErrNoDocuments {
				// Получаем информацию о кошельке
				nftItems, apiErr := h.TonApi.GetWalletNftItems(ctx, walletAddress)
				if apiErr != nil {
					return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Error while getting wallet nft items: %v", apiErr))
				}

				wallet := wallet.Wallet{
					Address:        walletAddress,
					NftItems:       *nftItems,
					NftCollections: nil,
				}

				if addErr := h.WalletDB.AddWallet(ctx, &wallet); addErr != nil {
					return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Error while adding wallet to DB: %v", addErr))
				}

				return c.Status(fiber.StatusCreated).JSON(wallet)
			}

			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Error while getting wallet data: %v", dbErr))
		}

		return c.Status(fiber.StatusOK).JSON(foundWallet)
	}
}

func (h *WalletHandler) RefreshWalletNftItems() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		walletAddress := c.Query("wallet-address")
		if walletAddress == "" {
			return c.Status(fiber.StatusBadRequest).SendString("Wallet Address is required")
		}

		nftItems, apiErr := h.TonApi.GetWalletNftItems(ctx, walletAddress)
		if apiErr != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Error while getting wallet nft items: %v", apiErr))
		}

		if updateErr := h.WalletDB.RefreshWalletNftItems(ctx, walletAddress, nftItems); updateErr != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Error update NFT items: %v", updateErr))
		}

		return c.Status(fiber.StatusOK).JSON(nftItems)
	}
}
