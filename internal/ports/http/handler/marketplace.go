package handler

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/xssnick/tonutils-go/tlb"

	marketplacecontractservice "github.com/rom6n/create-nft-go/internal/service/marketplace_contract_service"
)

type MarketplaceContractHandler struct {
	MarketplaceContractService marketplacecontractservice.MarketplaceContractServiceRepository
}

func (v *MarketplaceContractHandler) DepositMarket() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		value := c.Query("amount")
		isTest := c.Query("is-testnet")

		if value == "" || isTest == "" {
			return c.Status(fiber.StatusBadRequest).SendString("amount and is testnet are required")
		}

		isTestnet, parseBoolErr := strconv.ParseBool(isTest)
		if parseBoolErr != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("error parsing is-testnet to bool: %v", parseBoolErr))
		}

		amount, parseUintErr := strconv.ParseUint(value, 0, 64)
		if parseUintErr != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("error parsing amount to uint64: %v", parseUintErr))
		}

		depositErr := v.MarketplaceContractService.DepositMarketplaceContract(ctx, amount, isTestnet)
		if depositErr != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("error depositing marketplace contract: %v", depositErr))
		}

		return c.Status(fiber.StatusOK).SendString(fmt.Sprintf("Successfully deposited %v TON", tlb.FromNanoTONU(amount)))
	}
}

func (v *MarketplaceContractHandler) DeployMarketContract() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		isTest := c.Query("is-testnet")
		if isTest == "" {
			return c.Status(fiber.StatusBadRequest).SendString("is testnet is required")
		}

		isTestnet, parseBoolErr := strconv.ParseBool(isTest)
		if parseBoolErr != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("error parsing is-testnet to bool: %v", parseBoolErr))
		}

		deployErr := v.MarketplaceContractService.DeployMarketplaceContract(ctx, isTestnet)
		if deployErr != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("error deploying marketplace contract: %v", deployErr))
		}

		return c.Status(fiber.StatusOK).SendString("Successfully deployed!")
	}
}
