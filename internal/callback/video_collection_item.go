package callback

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/avast/retry-go/v4"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
)

var TypeVideoCollectionItem = "video_collection_item"

type VideoCollectionItem struct {
	ChatID    int64
	MessageID int
	Episode   int32
}

func (vci VideoCollectionItem) GenInlineKeyboard(inlineKeyboardType string) tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(tg.NewInlineKeyboardRow(
		tg.NewInlineKeyboardButtonData("Buat Tautan", fmt.Sprintf("%s:%d", inlineKeyboardType, vci.Episode)),
		tg.NewInlineKeyboardButtonData("Tidak", fmt.Sprintf("%s:%s", TypeVideoCollection, "")),
	))
}

func VideoCollectionItemHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	logger := log.Ctx(ctx).With().Logger()
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	episode, err := strconv.Atoi(strings.Split(update.CallbackQuery.Data, ":")[1])
	if err != nil {
		logger.Err(err).Msg("failed to convert episode string to int")
		return
	}

	vci := VideoCollectionItem{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Episode:   int32(episode),
	}

	text := fmt.Sprintf(`Tombol "Buat Tautan" akan membuat tautan untuk menonton video Running Man episode %d. Apakah kamu ingin membuat tautan?`, vci.Episode)
	_, err = retry.DoWithData(
		func() (*models.Message, error) {
			return b.EditMessageText(ctx, &bot.EditMessageTextParams{
				ChatID:      vci.ChatID,
				MessageID:   vci.MessageID,
				Text:        text,
				ReplyMarkup: vci.GenInlineKeyboard(TypeVideoLink),
			})
		},
		retry.Attempts(3),
		retry.LastErrorOnly(true),
	)

	if err != nil {
		logger.Err(err).Msgf("failed to send %s callback edit message", TypeVideoCollectionItem)
		return
	}
}
