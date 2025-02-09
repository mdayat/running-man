package command

import (
	"context"

	"github.com/avast/retry-go/v4"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/mdayat/running-man/internal/callback"
	"github.com/rs/zerolog/log"
)

func BrowseHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	logger := log.Ctx(ctx).With().Logger()
	rml := callback.RunningManLibraries{
		ChatID:    update.Message.Chat.ID,
		MessageID: update.Message.ID,
	}

	years, err := rml.GetRunningManYears(ctx)
	if err != nil {
		logger.Err(err).Send()
		return
	}
	rml.Years = years

	_, err = retry.DoWithData(
		func() (*models.Message, error) {
			msg := &bot.SendMessageParams{
				ChatID: rml.ChatID,
				Text:   "",
			}

			if len(years) == 0 {
				msg.Text = "Ooops... kami tidak memiliki daftar video Running Man untuk ditampilkan."
			} else {
				msg.Text = callback.LibrariesTextMsg
				msg.ReplyMarkup = rml.GenInlineKeyboard(callback.TypeVideoList)
			}

			return b.SendMessage(ctx, msg)
		},
		retry.Attempts(3),
		retry.LastErrorOnly(true),
	)

	if err != nil {
		logger.Err(err).Msg("failed to send browse command message")
		return
	}
}
