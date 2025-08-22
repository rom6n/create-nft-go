package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	convertimageservice "github.com/rom6n/create-nft-go/internal/service/convert_image"
)

type ConverterHandler struct {
	ConvertImageService convertimageservice.ConvertImageRepository
}

func (v *ConverterHandler) ConvertToWebP() fiber.Handler {
	return func(c *fiber.Ctx) error {

		img := c.Body()

		newImg, err := v.ConvertImageService.ConvertImageWebP(img)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("error converting: %v", err))
		}

		return c.Status(fiber.StatusOK).Send(newImg)
	}
}