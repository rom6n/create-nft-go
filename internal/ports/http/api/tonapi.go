package api

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/tonkeeper/tonapi-go"
)

func NewTonApiClient() *tonapi.Client {
	if loadErr := godotenv.Load(); loadErr != nil {
		log.Fatal("Error loading .env file")
	}

	token := os.Getenv("TONAPI_TOKEN")
	if token == "" {
		log.Fatal("Error. Add TonApi token to env.")
	}

	client, err := tonapi.NewClient(tonapi.TonApiURL, tonapi.WithToken(token))
	if err != nil {
		log.Fatalf("TonApi connection error: %v\n", err)
	}

	return client
}
