package command

import (
	"context"
	"fmt"

	"github.com/avast/retry-go/v4"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/mdayat/running-man/configs/env"
	"github.com/rs/zerolog/log"
)

func SupportHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	logger := log.Ctx(ctx).With().Logger()
	text := "Hai! Terima kasih telah menggunakan aplikasi kami.\n\nJika Anda membutuhkan bantuan teknis, tips, atau jawaban atas pertanyaan terkait penggunaan aplikasi, jangan ragu untuk menghubungi kami melalui tombol \"Narahubung\"."
	_, err := retry.DoWithData(
		func() (*models.Message, error) {
			return b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   text,
				ReplyMarkup: models.InlineKeyboardMarkup{
					InlineKeyboard: [][]models.InlineKeyboardButton{
						{
							{Text: "Narahubung", URL: fmt.Sprintf("https://wa.me/%s", env.SupportNumber)},
						},
					},
				},
			})
		},
		retry.Attempts(3),
		retry.LastErrorOnly(true),
	)

	if err != nil {
		logger.Err(err).Msg("failed to send support command message")
		return
	}
}
