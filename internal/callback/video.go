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

var TypeVideo = "video"

type RunningManVideo struct {
	ChatID    int64
	MessageID int
	Year      int32
	Episode   int32
}

func (rmv RunningManVideo) GenInlineKeyboard(inlineKeyboardType string) tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(tg.NewInlineKeyboardRow(
		tg.NewInlineKeyboardButtonData("Iya", fmt.Sprintf("%s:%d,%d", inlineKeyboardType, rmv.Year, rmv.Episode)),
		tg.NewInlineKeyboardButtonData("Tidak", fmt.Sprintf("%s:%d", TypeVideos, rmv.Year)),
	))
}

func VideoHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	logger := log.Ctx(ctx).With().Logger()
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	splittedData := strings.Split(update.CallbackQuery.Data, ":")[1]
	year, err := strconv.Atoi(strings.Split(splittedData, ",")[0])
	if err != nil {
		logger.Err(err).Msg("failed to convert year string to int")
		return
	}

	episode, err := strconv.Atoi(strings.Split(splittedData, ",")[1])
	if err != nil {
		logger.Err(err).Msg("failed to convert episode string to int")
		return
	}

	rmv := RunningManVideo{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Year:      int32(year),
		Episode:   int32(episode),
	}

	text := fmt.Sprintf("Apakah kamu ingin membeli Running Man episode %d?", rmv.Episode)
	_, err = retry.DoWithData(
		func() (*models.Message, error) {
			return b.EditMessageText(ctx, &bot.EditMessageTextParams{
				ChatID:      rmv.ChatID,
				MessageID:   rmv.MessageID,
				Text:        text,
				ReplyMarkup: rmv.GenInlineKeyboard(TypeInvoice),
			})
		},
		retry.Attempts(3),
		retry.LastErrorOnly(true),
	)

	if err != nil {
		logger.Err(err).Msgf("failed to send %s callback edit message", TypeVideo)
		return
	}
}
