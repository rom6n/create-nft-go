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

func (v *MarketplaceContractHandler) WithdrawTonFromMarketContract() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		isTest, text, value := c.Query("is-testnet"), c.Query("message"), c.Query("amount")
		if isTest == "" || value == "" {
			return c.Status(fiber.StatusBadRequest).SendString("is-testnet and amount queries are required")
		}

		var message []string
		if text != "" {
			message = append(message, text)
		}

		amount, parseUintErr := strconv.ParseUint(value, 0, 64)
		if parseUintErr != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("error parsing amount to uint64: %v", parseUintErr))
		}

		isTestnet, parseBoolErr := strconv.ParseBool(isTest)
		if parseBoolErr != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("error parsing is-testnet to bool: %v", parseBoolErr))
		}

		withdrawErr := v.MarketplaceContractService.WithdrawTonFromMarketplaceContract(ctx, amount, isTestnet, message...)
		if withdrawErr != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("error withdrawing ton from marketplace contract: %v", withdrawErr))
		}

		return c.Status(fiber.StatusOK).SendString(fmt.Sprintf("Successfully withdrawed %v TON\n", tlb.FromNanoTONU(amount)))
	}
}
