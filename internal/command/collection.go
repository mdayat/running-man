package command

import (
	"context"

	"github.com/avast/retry-go/v4"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/mdayat/running-man/internal/callback"
	"github.com/rs/zerolog/log"
)

func CollectionHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	logger := log.Ctx(ctx).With().Logger()
	vc := callback.VideoCollection{
		ChatID: update.Message.Chat.ID,
		UserID: update.Message.From.ID,
	}

	episodes, err := vc.GetEpisodesFromUserVideoCollection()
	if err != nil {
		logger.Err(err).Send()
		return
	}
	vc.Episodes = episodes

	_, err = retry.DoWithData(
		func() (*models.Message, error) {
			return b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      update.Message.Chat.ID,
				Text:        callback.VideoCollectionTextMsg,
				ParseMode:   models.ParseModeHTML,
				ReplyMarkup: vc.GenInlineKeyboard(callback.TypeVideoCollectionDetail),
			})
		},
		retry.Attempts(3),
		retry.LastErrorOnly(true),
	)

	if err != nil {
		logger.Err(err).Msg("failed to send collection command message")
		return
	}
}
