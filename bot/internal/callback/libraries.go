package callback

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/avast/retry-go/v4"
	badger "github.com/dgraph-io/badger/v4"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/mdayat/running-man/configs/services"
	"github.com/mdayat/running-man/internal/converter"
	"github.com/rs/zerolog/log"
)

var (
	TypeLibraries    = "libraries"
	LibrariesTextMsg = "Pilih tahun Running Man:"
)

type RunningManLibraries struct {
	ChatID    int64
	MessageID int
	Years     []int32
}

func (rml RunningManLibraries) GetRunningManYears(ctx context.Context) ([]int32, error) {
	var years []int32
	err := services.Badger.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(TypeLibraries))
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return fmt.Errorf("failed to get %s key: %w", TypeLibraries, err)
		}

		if err != nil && errors.Is(err, badger.ErrKeyNotFound) {
			retryFunc := func() ([]int32, error) {
				return services.Queries.GetRunningManYears(ctx)
			}

			years, err = retry.DoWithData(retryFunc, retry.Attempts(3), retry.LastErrorOnly(true))
			if err != nil {
				return fmt.Errorf("failed to get running man years: %w", err)
			}

			entryVal, err := converter.Int32SliceToBytes(years)
			if err != nil {
				return fmt.Errorf("failed to convert int32 of years to bytes: %w", err)
			}

			entry := badger.NewEntry([]byte(TypeLibraries), entryVal).WithTTL(time.Hour * 24)
			if err := txn.SetEntry(entry); err != nil {
				return fmt.Errorf("failed to set %s key: %w", TypeLibraries, err)
			}

			return nil
		}

		valCopy, err := item.ValueCopy(nil)
		if err != nil {
			return fmt.Errorf("failed to copy the value of %s key: %w", TypeLibraries, err)
		}

		years, err = converter.BytesToInt32Slice(valCopy)
		if err != nil {
			return fmt.Errorf("failed to convert bytes of valCopy to int32: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to execute badger update function: %w", err)
	}

	return years, nil
}

func (rml RunningManLibraries) GenInlineKeyboard(inlineKeyboardType string) models.InlineKeyboardMarkup {
	numOfRowItems := 3
	numOfRows := int(math.Ceil(float64(len(rml.Years)) / float64(numOfRowItems)))

	inlineKeyboardRows := make([][]models.InlineKeyboardButton, 0, numOfRows)
	inlineKeyboardRowItems := make([]models.InlineKeyboardButton, 0, numOfRowItems)

	for _, v := range rml.Years {
		btnText := fmt.Sprintf("%d", v)
		btnData := fmt.Sprintf("%s:%d", inlineKeyboardType, v)
		inlineKeyboardRowItems = append(inlineKeyboardRowItems, models.InlineKeyboardButton{Text: btnText, CallbackData: btnData})

		if len(inlineKeyboardRowItems) == numOfRowItems {
			inlineKeyboardRows = append(inlineKeyboardRows, inlineKeyboardRowItems)
			inlineKeyboardRowItems = make([]models.InlineKeyboardButton, 0, numOfRowItems)
		}
	}

	if len(inlineKeyboardRowItems) != 0 {
		inlineKeyboardRows = append(inlineKeyboardRows, inlineKeyboardRowItems)
	}

	return models.InlineKeyboardMarkup{InlineKeyboard: inlineKeyboardRows}
}

func LibrariesHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	logger := log.Ctx(ctx).With().Logger()
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	rml := RunningManLibraries{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
	}

	years, err := rml.GetRunningManYears(ctx)
	if err != nil {
		logger.Err(err).Send()
		return
	}
	rml.Years = years

	_, err = retry.DoWithData(
		func() (*models.Message, error) {
			return b.EditMessageText(ctx, &bot.EditMessageTextParams{
				ChatID:      rml.ChatID,
				MessageID:   rml.MessageID,
				Text:        LibrariesTextMsg,
				ReplyMarkup: rml.GenInlineKeyboard(TypeVideoList),
			})
		},
		retry.Attempts(3),
		retry.LastErrorOnly(true),
	)

	if err != nil {
		logger.Err(err).Msgf("failed to send %s callback edit message", TypeLibraries)
		return
	}
}
