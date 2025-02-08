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
	TripayMerchantCode string
	TripayAPIKey       string
	TripayPrivateKey   string
	TripayURL          string
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
	TripayMerchantCode = os.Getenv("TRIPAY_MERCHANT_CODE")
	TripayAPIKey = os.Getenv("TRIPAY_API_KEY")
	TripayPrivateKey = os.Getenv("TRIPAY_PRIVATE_KEY")
	TripayURL = os.Getenv("TRIPAY_URL")

	return nil
}
