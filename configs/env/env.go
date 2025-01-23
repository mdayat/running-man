package env

import (
	"os"

	"github.com/joho/godotenv"
)

var (
	BotToken           string
	DatabaseURL        string
	DirectEmbedBaseURL string
	SupportNumber      string
)

func Load() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}

	BotToken = os.Getenv("BOT_TOKEN")
	DatabaseURL = os.Getenv("DATABASE_URL")
	DirectEmbedBaseURL = os.Getenv("DIRECT_EMBED_BASE_URL")
	SupportNumber = os.Getenv("SUPPORT_NUMBER")

	return nil
}
