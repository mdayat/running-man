package env

import (
	"os"

	"github.com/joho/godotenv"
)

var (
	BOT_TOKEN    string
	DATABASE_URL string
)

func New() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}

	BOT_TOKEN = os.Getenv("BOT_TOKEN")
	DATABASE_URL = os.Getenv("DATABASE_URL")

	return nil
}
