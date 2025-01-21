package env

import (
	"os"

	"github.com/joho/godotenv"
)

var (
	BOT_TOKEN             string
	DATABASE_URL          string
	TOKEN_AUTH_KEY_2015   string
	TOKEN_AUTH_KEY_2018   string
	TOKEN_AUTH_KEY_2020   string
	TOKEN_AUTH_KEY_2021   string
	DIRECT_EMBED_BASE_URL string
)

func Load() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}

	BOT_TOKEN = os.Getenv("BOT_TOKEN")
	DATABASE_URL = os.Getenv("DATABASE_URL")
	TOKEN_AUTH_KEY_2015 = os.Getenv("TOKEN_AUTH_KEY_2015")
	TOKEN_AUTH_KEY_2018 = os.Getenv("TOKEN_AUTH_KEY_2018")
	TOKEN_AUTH_KEY_2020 = os.Getenv("TOKEN_AUTH_KEY_2020")
	TOKEN_AUTH_KEY_2021 = os.Getenv("TOKEN_AUTH_KEY_2021")
	DIRECT_EMBED_BASE_URL = os.Getenv("DIRECT_EMBED_BASE_URL")

	return nil
}
