package main

import (
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/rom6n/create-nft-go/internal/database/nosql"
	"github.com/rom6n/create-nft-go/internal/ports/http/api"
	"github.com/rom6n/create-nft-go/internal/ports/http/handler"
)

func main() {
	mongoClient := nosql.NewMongoClient()
	tonApiClient := api.NewTonApiClient()

	app := fiber.New(fiber.Config{
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})

	app.Use(logger.New())

	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("pong")
	})

	app.Get("/favicon.ico", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	api := app.Group("/api")
	api.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("pong")
	})

	api.Get("/get-wallet-data", handler.GetWalletData(mongoClient, tonApiClient))

	api.Get("/update-wallet-nft-items", handler.UpdateWalletNftItems(mongoClient, tonApiClient)) // В будущем поменять на PUT

	app.Listen(":2000")
}
