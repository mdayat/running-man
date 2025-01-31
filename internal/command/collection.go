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
		ChatID:    update.Message.Chat.ID,
		UserID:    update.Message.From.ID,
		MessageID: update.Message.ID,
	}

	episodes, err := vc.GetEpisodesFromUserVideoCollection(ctx)
	if err != nil {
		logger.Err(err).Send()
		return
	}
	vc.Episodes = episodes

	_, err = retry.DoWithData(
		func() (*models.Message, error) {
			msg := &bot.SendMessageParams{
				ChatID: vc.ChatID,
				Text:   "",
			}

			if len(episodes) == 0 {
				msg.Text = "Ooops... kamu belum memiliki video Running Man! Yuk jelajahi dan lakukan pembelian video Running Man melalui perintah /browse."
			} else {
				msg.Text = callback.VideoCollectionTextMsg
				msg.ReplyMarkup = vc.GenInlineKeyboard(callback.TypeVideoCollectionItem)
			}

			return b.SendMessage(ctx, msg)
		},
		retry.Attempts(3),
		retry.LastErrorOnly(true),
	)

	if err != nil {
		logger.Err(err).Msg("failed to send collection command message")
		return
	}
}
