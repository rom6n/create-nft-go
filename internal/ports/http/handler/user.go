package handler

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/rom6n/create-nft-go/internal/domain/user"
)

type UserHandler struct {
	UserDB user.UserRepository
}

func (h *UserHandler) GetUserData() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		userStrID := c.Query("user-id")
		if userStrID == "" {
			return c.Status(fiber.StatusBadRequest).SendString("User ID is required")
		}

		userID, parseErr := strconv.ParseInt(userStrID, 0, 64)
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).SendString("User ID must be an int")
		}

		user, dbErr := h.UserDB.GetUserByID(ctx, userID)
		if dbErr != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Error while getting user data: %v", dbErr))
		}

		return c.Status(fiber.StatusOK).JSON(user)
	}
}
