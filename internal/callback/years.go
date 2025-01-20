package callback

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/avast/retry-go/v4"
	badger "github.com/dgraph-io/badger/v4"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mdayat/running-man/configs/services"
	"github.com/mdayat/running-man/internal/converter"
)

var (
	TypeYears    InlineKeyboardType = "years"
	YearsTextMsg                    = "Pilih tahun Running Man:"
)

type RunningManYears struct {
	ChatID    int64
	MessageID int
	Years     []int32
}

func (rml RunningManYears) GetRunningManYears() ([]int32, error) {
	var years []int32
	err := services.Badger.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(TypeYears))
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return fmt.Errorf("failed to get running man years key: %w", err)
		}

		if err != nil && errors.Is(err, badger.ErrKeyNotFound) {
			retryFunc := func() ([]int32, error) {
				return services.Queries.GetRunningManLibraries(context.TODO())
			}

			years, err = retry.DoWithData(retryFunc, retry.Attempts(3))
			if err != nil {
				return fmt.Errorf("failed to get running man years: %w", err)
			}

			entryVal, err := converter.Int32SliceToBytes(years)
			if err != nil {
				return fmt.Errorf("failed to convert int32 of years to bytes: %w", err)
			}

			entry := badger.NewEntry([]byte(TypeYears), entryVal).WithTTL(time.Hour)
			if err := txn.SetEntry(entry); err != nil {
				return fmt.Errorf("failed to set running man years key: %w", err)
			}

			return nil
		}

		valCopy, err := item.ValueCopy(nil)
		if err != nil {
			return fmt.Errorf("failed to copy the value of years key: %w", err)
		}

		years, err = converter.BytesToInt32Slice(valCopy)
		if err != nil {
			return fmt.Errorf("failed to convert bytes of valCopy to int32: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return years, nil
}

func (rml RunningManYears) GenInlineKeyboard(inlineKeyboardType InlineKeyboardType) tg.InlineKeyboardMarkup {
	numOfRowItems := 3
	numOfRows := int(math.Ceil(float64(len(rml.Years) / numOfRowItems)))

	inlineKeyboardRows := make([][]tg.InlineKeyboardButton, 0, numOfRows)
	inlineKeyboardRowItems := make([]tg.InlineKeyboardButton, 0, numOfRowItems)

	for _, v := range rml.Years {
		btnText := fmt.Sprintf("%d", v)
		btnData := fmt.Sprintf("%s:%d", inlineKeyboardType, v)
		inlineKeyboardRowItems = append(inlineKeyboardRowItems, tg.NewInlineKeyboardButtonData(btnText, btnData))

		if len(inlineKeyboardRowItems) == numOfRowItems {
			inlineKeyboardRows = append(inlineKeyboardRows, tg.NewInlineKeyboardRow(inlineKeyboardRowItems...))
			inlineKeyboardRowItems = inlineKeyboardRowItems[:0]
		}
	}

	if len(inlineKeyboardRowItems) != 0 {
		inlineKeyboardRows = append(inlineKeyboardRows, tg.NewInlineKeyboardRow(inlineKeyboardRowItems...))
	}

	return tg.NewInlineKeyboardMarkup(inlineKeyboardRows...)
}

func (rml RunningManYears) Process() (tg.Chattable, error) {
	years, err := rml.GetRunningManYears()
	if err != nil {
		return nil, err
	}
	rml.Years = years

	chat := tg.NewEditMessageTextAndMarkup(rml.ChatID, rml.MessageID, YearsTextMsg, rml.GenInlineKeyboard(TypeEpisodes))
	return chat, nil
}
