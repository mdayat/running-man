package main

import (
	"path/filepath"
	"strconv"

	"github.com/mdayat/movies-telegram-bot/configs/env"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}

	logger := log.With().Caller().Logger()
	err := env.Init()
	if err != nil {
		logger.Fatal().Err(err).Send()
	}
}
