package commands

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var defaultMessage = `
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

type DefaultCommand struct {
	ChatID int64
}

func (dc DefaultCommand) Process() (msg tgbotapi.MessageConfig, _ error) {
	msg = tgbotapi.NewMessage(dc.ChatID, defaultMessage)
	return msg, nil
}
