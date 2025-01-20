package main

import (
	"context"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/avast/retry-go/v4"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mdayat/running-man/configs/env"
	"github.com/mdayat/running-man/configs/services"
	"github.com/mdayat/running-man/internal/callback"
	"github.com/mdayat/running-man/internal/command"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func sendChat(bot *tg.BotAPI, chat tg.Chattable) error {
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

	bot, err := tg.NewBotAPI(env.BOT_TOKEN)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	updateConfig := tg.NewUpdate(0)
	updateConfig.Timeout = 60
	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message != nil && !update.Message.IsCommand() {
			continue
		}

		if update.Message != nil {
			switch update.Message.Command() {
			case "browse":
				bc := command.Browse{ChatID: update.Message.Chat.ID}
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
				dc := command.Default{
					UserID:    update.Message.From.ID,
					ChatID:    update.Message.Chat.ID,
					FirstName: update.Message.From.FirstName,
				}

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
			continue
		}

		if update.CallbackQuery != nil {
			splittedCallbackData := strings.Split(update.CallbackQuery.Data, ":")
			callbackType := callback.InlineKeyboardType(splittedCallbackData[0])

			switch callbackType {
			case callback.TypeYears:
				rml := callback.RunningManYears{
					ChatID:    update.CallbackQuery.Message.Chat.ID,
					MessageID: update.CallbackQuery.Message.MessageID,
				}

				chat, err := rml.Process()
				if err != nil {
					logger.Err(err).Msg("failed to process running man years callback")
					continue
				}

				if err := sendChat(bot, chat); err != nil {
					logger.Err(err).Msg("failed to send updated chat for running man years inline keyboard")
					continue
				}
			case callback.TypeEpisodes:
				year, err := strconv.Atoi(splittedCallbackData[1])
				if err != nil {
					logger.Err(err).Msg("failed to convert running man year string to int")
					continue
				}

				rme := callback.RunningManEpisodes{
					Year:      int32(year),
					ChatID:    update.CallbackQuery.Message.Chat.ID,
					MessageID: update.CallbackQuery.Message.MessageID,
				}

				chat, err := rme.Process()
				if err != nil {
					logger.Err(err).Msg("failed to process running man episodes callback")
					continue
				}

				if err := sendChat(bot, chat); err != nil {
					logger.Err(err).Msg("failed to send updated chat for running man episodes inline keyboard")
					continue
				}
			case callback.TypeEpisode, callback.TypePurchase:
				year, err := strconv.Atoi(splittedCallbackData[1])
				if err != nil {
					logger.Err(err).Msg("failed to convert running man year string to int")
					continue
				}

				episode, err := strconv.Atoi(splittedCallbackData[2])
				if err != nil {
					logger.Err(err).Msg("failed to convert running man episode string to int")
					continue
				}

				rme := callback.RunningManEpisode{
					UserID:       update.CallbackQuery.From.ID,
					ChatID:       update.CallbackQuery.Message.Chat.ID,
					MessageID:    update.CallbackQuery.Message.MessageID,
					Year:         int32(year),
					Episode:      int32(episode),
					IsPurchasing: callbackType == callback.TypePurchase,
				}

				chat, err := rme.Process()
				if err != nil {
					logger.Err(err).Msg("failed to process running man episode callback")
					continue
				}

				if err := sendChat(bot, chat); err != nil {
					logger.Err(err).Msg("failed to send updated chat for running man episode inline keyboard")
					continue
				}
			default:
			}
			continue
		}
	}
}
