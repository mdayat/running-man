package main

import (
	"path/filepath"
	"strconv"

	"github.com/avast/retry-go/v4"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mdayat/running-man/commands"
	"github.com/mdayat/running-man/configs/env"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func sendChat(bot *tgbotapi.BotAPI, chat tgbotapi.Chattable) error {
	retryFunc := func() error {
		if _, err := bot.Send(chat); err != nil {
			return err
		}

		return nil
	}

	if err := retry.Do(retryFunc, retry.Attempts(3)); err != nil {
		return err
	}

	return nil
}

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

	bot, err := tgbotapi.NewBotAPI(env.BOT_TOKEN)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if !update.Message.IsCommand() {
			continue
		}

		switch update.Message.Command() {
		case "browse":
			bc := commands.Browse{ChatID: update.Message.Chat.ID}
			chat, err := bc.Process()
			if err != nil {
				logger.Err(err).Msg("failed to process browse command")
				continue
			}

			if err := sendChat(bot, chat); err != nil {
				logger.Err(err).Msg("failed to send browse command's chat")
				continue
			}
		case "start", "help":
			fallthrough
		default:
			dc := commands.Default{ChatID: update.Message.Chat.ID}
			chat, err := dc.Process()
			if err != nil {
				logger.Err(err).Msg("failed to process default command")
				continue
			}

			if err := sendChat(bot, chat); err != nil {
				logger.Err(err).Msg("failed to send default command's chat")
				continue
			}
		}
	}
}
