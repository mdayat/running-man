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
		ChatID: update.Message.Chat.ID,
	}

	years, err := rml.GetRunningManYears()
	if err != nil {
		logger.Err(err).Send()
		return
	}
	rml.Years = years

	_, err = retry.DoWithData(
		func() (*models.Message, error) {
			return b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      update.Message.Chat.ID,
				Text:        callback.LibrariesTextMsg,
				ParseMode:   models.ParseModeHTML,
				ReplyMarkup: rml.GenInlineKeyboard(callback.TypeVideos),
			})
		},
		retry.Attempts(3),
		retry.LastErrorOnly(true),
	)

	if err != nil {
		logger.Err(err).Msg("failed to send browse command message")
		return
	}
}
