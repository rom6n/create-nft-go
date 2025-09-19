package telegutils

import (
	"log"
	"os"
	"time"

	initdata "github.com/telegram-mini-apps/init-data-golang"
)

func VerifyTelegramInitData(initString string, botToken string) bool {
	if err := initdata.Validate(initString, botToken, 24*time.Hour); err != nil {
		log.Printf("Not authorized login try: %v \n", err)
		return false
	}

	return true
}

func GetBotToken() string {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatalf("TELEGRAM_BOT_TOKEN env var not set \n")
	}

	return token
}
