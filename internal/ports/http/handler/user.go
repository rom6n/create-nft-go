package handler

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	userservice "github.com/rom6n/create-nft-go/internal/service/user_service"
)

type UserHandler struct {
	UserService userservice.UserServiceRepository
}

func (h *UserHandler) GetUserData() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		userStrID := c.Params("id")

		userID, parseErr := strconv.ParseInt(userStrID, 0, 64)
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).SendString("User ID must be an int")
		}

		user, dbErr := h.UserService.GetUserByID(ctx, userID)
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
