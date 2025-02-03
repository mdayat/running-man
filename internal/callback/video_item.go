package callback

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/avast/retry-go/v4"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/mdayat/running-man/configs/services"
	"github.com/rs/zerolog/log"
)

var TypeVideoItem = "video_item"

type VideoItem struct {
	ChatID    int64
	UserID    int64
	MessageID int
	Year      int32
	Episode   int32
}

func (vi VideoItem) GenInlineKeyboard(inlineKeyboardType string) models.InlineKeyboardMarkup {
	return models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Iya", CallbackData: fmt.Sprintf("%s:%d,%d", inlineKeyboardType, vi.Year, vi.Episode)},
				{Text: "Tidak", CallbackData: fmt.Sprintf("%s:%d", TypeVideoList, vi.Year)},
			},
		},
	}
}

func (vi VideoItem) HandleUnsubscribedUser() *bot.EditMessageTextParams {
	text := fmt.Sprintf(
		"Kamu tidak bisa menonton video Running Man episode %d karena kamu belum berlangganan.\n\nKlik tombol \"Berlangganan\" agar kamu bisa menonton video Running Man sepuasnya.",
		vi.Episode,
	)

	inlineKeyboard := models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Berlangganan", CallbackData: "TODO"},
				{Text: "Kembali", CallbackData: fmt.Sprintf("%s:%d", TypeVideoList, vi.Year)},
			},
		},
	}

	msg := &bot.EditMessageTextParams{
		ChatID:      vi.ChatID,
		MessageID:   vi.MessageID,
		Text:        text,
		ReplyMarkup: inlineKeyboard,
	}

	return msg
}

func (vi VideoItem) HandleSubscribedUser() *bot.EditMessageTextParams {
	text := fmt.Sprintf(
		"Klik tombol \"Tonton\" untuk menonton video Running Man episode %d. Kamu akan mendapatkan tautan untuk menonton video tersebut.",
		vi.Episode,
	)

	inlineKeyboard := models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Tonton", CallbackData: fmt.Sprintf("%s:%d", TypeVideoLink, vi.Episode)},
				{Text: "Kembali", CallbackData: fmt.Sprintf("%s:%d", TypeVideoList, vi.Year)},
			},
		},
	}

	msg := &bot.EditMessageTextParams{
		ChatID:      vi.ChatID,
		MessageID:   vi.MessageID,
		Text:        text,
		ReplyMarkup: inlineKeyboard,
	}

	return msg
}

func VideoItemHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
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

	vi := VideoItem{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		UserID:    update.CallbackQuery.From.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Year:      int32(year),
		Episode:   int32(episode),
	}

	isUserSubscribed, err := retry.DoWithData(
		func() (bool, error) {
			return services.Queries.IsUserSubscribed(ctx, vi.UserID)
		},
		retry.Attempts(3),
		retry.LastErrorOnly(true),
	)

	if err != nil {
		logger.Err(err).Msg("failed to check if user already subscribed or not")
		return
	}

	var msg *bot.EditMessageTextParams
	if isUserSubscribed {
		msg = vi.HandleSubscribedUser()
	} else {
		msg = vi.HandleUnsubscribedUser()
	}

	_, err = retry.DoWithData(
		func() (*models.Message, error) {
			return b.EditMessageText(ctx, msg)
		},
		retry.Attempts(3),
		retry.LastErrorOnly(true),
	)

	if err != nil {
		logger.Err(err).Msgf("failed to send %s callback edit message", TypeVideoItem)
		return
	}
}
