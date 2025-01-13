package main

import (
	"path/filepath"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mdayat/running-man/configs/env"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var DefaultMessage = `
Selamat datang di @RunningManSeriesBot! Dengan bot ini, Anda dapat mengelola dan membeli episode lama dari "Running Man." Berikut adalah perintah yang dapat Anda gunakan:

/start - Selamat datang dan pengenalan
/help - Dapatkan bantuan dan daftar perintah yang tersedia
/browse - Tampilkan katalog episode
/search [nomor_episode] - Cari episode tertentu
/buy [nomor_episode] - Beli episode
/checkout - Selesaikan pembelian dengan Telegram stars
/mycollection - Tampilkan episode yang sudah dibeli
/feedback - Kirimkan umpan balik atau laporkan masalah
`

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

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		switch update.Message.Command() {
		case "start":
			msg.Text = DefaultMessage
		case "help":
			msg.Text = DefaultMessage
		default:
			msg.Text = DefaultMessage
		}

		if _, err := bot.Send(msg); err != nil {
			logger.Err(err).Msg("failed to send message")
			continue
		}
	}
}
