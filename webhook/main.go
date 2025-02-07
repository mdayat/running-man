package main

import (
	"context"
	"net/http"
	"path/filepath"
	"strconv"
	"webhook/configs/env"
	"webhook/configs/services"
	"webhook/internal"

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

	app := internal.NewApp()
	if err := http.ListenAndServe(":8080", app); err != nil {
		logger.Fatal().Err(err).Msg("failed to serve HTTP server")
	}
}
