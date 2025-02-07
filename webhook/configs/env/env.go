package env

import (
	"os"

	"github.com/joho/godotenv"
)

var (
	DatabaseURL        string
	TripayMerchantCode string
	TripayAPIKey       string
	TripayPrivateKey   string
	AllowedOrigins     string
	BotToken           string
)

func Load() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}

	DatabaseURL = os.Getenv("DATABASE_URL")
	TripayMerchantCode = os.Getenv("TRIPAY_MERCHANT_CODE")
	TripayAPIKey = os.Getenv("TRIPAY_API_KEY")
	TripayPrivateKey = os.Getenv("TRIPAY_PRIVATE_KEY")
	AllowedOrigins = os.Getenv("ALLOWED_ORIGINS")
	BotToken = os.Getenv("BOT_TOKEN")

	return nil
}
