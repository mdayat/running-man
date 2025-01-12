package env

import (
	"os"

	"github.com/joho/godotenv"
)

var (
	BOT_TOKEN string
)

func Init() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}

	BOT_TOKEN = os.Getenv("BOT_TOKEN")

	return nil
}
