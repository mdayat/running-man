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
	TypeLibraries    InlineKeyboardType = "libraries"
	LibrariesTextMsg                    = "Pilih tahun Running Man:"
)

type RunningManLibraries struct {
	ChatID    int64
	MessageID int
	Years     []int32
}

func (rml RunningManLibraries) GetRunningManYears() ([]int32, error) {
	var years []int32
	err := services.Badger.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(TypeLibraries))
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return fmt.Errorf("failed to get %s key: %w", TypeLibraries, err)
		}

		if err != nil && errors.Is(err, badger.ErrKeyNotFound) {
			retryFunc := func() ([]int32, error) {
				return services.Queries.GetRunningManYears(context.TODO())
			}

			years, err = retry.DoWithData(retryFunc, retry.Attempts(3))
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
		return nil, err
	}

	return years, nil
}

func (rml RunningManLibraries) GenInlineKeyboard(inlineKeyboardType InlineKeyboardType) tg.InlineKeyboardMarkup {
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

func (rml RunningManLibraries) Process() (tg.Chattable, error) {
	years, err := rml.GetRunningManYears()
	if err != nil {
		return nil, err
	}
	rml.Years = years

	chat := tg.NewEditMessageTextAndMarkup(rml.ChatID, rml.MessageID, LibrariesTextMsg, rml.GenInlineKeyboard(TypeEpisodes))
	return chat, nil
}
