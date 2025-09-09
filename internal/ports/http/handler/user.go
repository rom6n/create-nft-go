package handler

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	userservice "github.com/rom6n/create-nft-go/internal/service/user_service"
	"github.com/rom6n/create-nft-go/internal/service/withdraw_user_ton"
	"github.com/xssnick/tonutils-go/address"
)

type UserHandler struct {
	UserService         userservice.UserServiceRepository
	WithdrawUserService withdraw_user_ton.WithdrawUserTonRepository
}

func (v *UserHandler) GetUserData() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		userStrID := c.Params("id")

		userID, parseErr := strconv.ParseInt(userStrID, 0, 64)
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).SendString("User ID must be an int")
		}

		user, dbErr := v.UserService.GetUserByID(ctx, userID)
		if dbErr != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Error while getting user data: %v", dbErr))
		}

		return c.Status(fiber.StatusOK).JSON(user)
	}
}

func (v *UserHandler) GetUserNftCollections() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		userStrID := c.Params("id")

		userID, parseErr := strconv.ParseInt(userStrID, 0, 64)
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).SendString("User ID must be an int")
		}

		nftCollections := v.UserService.GetUserNftCollections(ctx, userID)

		return c.Status(fiber.StatusOK).JSON(nftCollections)
	}
}

func (v *UserHandler) GetUserNftItems() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		userStrID := c.Params("id")

		userID, parseErr := strconv.ParseInt(userStrID, 0, 64)
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).SendString("User ID must be an int")
		}

		nftItems := v.UserService.GetUserNftItems(ctx, userID)

		return c.Status(fiber.StatusOK).JSON(nftItems)
	}
}

func (v *UserHandler) WithdrawUserTON() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		userStrID := c.Params("id")
		withdrawTo := c.Query("withdraw-to")
		amountStr := c.Query("amount")
		isTest := c.Query("is-testnet")

		if withdrawTo == "" || isTest == "" || amountStr == "" {
			return c.Status(fiber.StatusBadRequest).SendString("withdraw-to, amount, is-testnet are required")
		}

		userID, parseErr := strconv.ParseInt(userStrID, 0, 64)
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("Error while parsing user ID: %v", parseErr))
		}

		withdrawAddress, addrParseErr := address.ParseAddr(withdrawTo)
		if addrParseErr != nil {
			return c.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("Error while parsing address: %v", addrParseErr))
		}

		amount, parseUintErr := strconv.ParseUint(amountStr, 0, 64)
		if parseUintErr != nil {
			return c.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("Error while parsing amount: %v", parseUintErr))
		}
		if amount <= 0 {
			return c.Status(fiber.StatusBadRequest).SendString("amount must be greater than zero")
		}

		isTestnet, parseBoolErr := strconv.ParseBool(isTest)
		if parseBoolErr != nil {
			return c.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("Error while parsing is-testnet: %v", parseBoolErr))
		}

		if withdrawErr := v.WithdrawUserService.Withdraw(ctx, userID, amount, withdrawAddress, isTestnet); withdrawErr != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Error while withdrawing: %v", withdrawErr))
		}

		return c.Status(fiber.StatusOK).SendString(fmt.Sprintf("Successfully withdrawed %v TON", amount/1000000000.0))
	}
}
