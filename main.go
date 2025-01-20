package main

import (
	"context"
	"path/filepath"
	"strconv"

	"github.com/mdayat/running-man/configs/env"
	"github.com/mdayat/running-man/configs/services"
	"github.com/mdayat/running-man/internal/bot"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}

	logger := log.With().Caller().Logger()
	if err := env.New(); err != nil {
		logger.Fatal().Err(err).Send()
	}

	ctx := context.TODO()
	db, err := services.NewDB(ctx, env.DATABASE_URL)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}
	defer db.Close()

	badger, err := services.NewBadger()
	if err != nil {
		logger.Fatal().Err(err).Send()
	}
	defer badger.Close()

	bot, err := bot.New()
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	bot.Logger = logger
	bot.Run()
}
