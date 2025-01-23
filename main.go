package main

import (
	"context"
	"os"
	"os/signal"
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
	zerolog.CallerMarshalFunc = func(_ uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}

	logger := log.With().Caller().Logger()
	if err := env.Load(); err != nil {
		logger.Fatal().Err(err).Msg("failed to load environment variables")
	}

	ctx := context.TODO()
	db, err := services.NewDB(ctx, env.DatabaseURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create database instance")
	}
	defer db.Close()

	badger, err := services.NewBadger()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create badger instance")
	}
	defer badger.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	bot, err := bot.New(env.BotToken)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create bot instance")
	}
	bot.Start(logger.WithContext(ctx))
}
