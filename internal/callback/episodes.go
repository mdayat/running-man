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
	TypeEpisodes    InlineKeyboardType = "episodes"
	EpisodesTextMsg                    = "Pilih episode Running Man:"
)

type RunningManEpisodes struct {
	Year      int32
	ChatID    int64
	MessageID int
	Episodes  []int32
}

func (rme RunningManEpisodes) GetRunningManEpisodes() ([]int32, error) {
	var episodes []int32
	err := services.Badger.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(fmt.Sprintf("%d:%s", rme.Year, TypeEpisodes)))
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return fmt.Errorf("failed to get running man episodes key: %w", err)
		}

		if err != nil && errors.Is(err, badger.ErrKeyNotFound) {
			retryFunc := func() ([]int32, error) {
				return services.Queries.GetRunningManVideosByYear(context.TODO(), rme.Year)
			}

			episodes, err = retry.DoWithData(retryFunc, retry.Attempts(3))
			if err != nil {
				return fmt.Errorf("failed to get running man episodes: %w", err)
			}

			entryVal, err := converter.Int32SliceToBytes(episodes)
			if err != nil {
				return fmt.Errorf("failed to convert int32 of episodes to bytes: %w", err)
			}

			entry := badger.NewEntry([]byte(fmt.Sprintf("%d:%s", rme.Year, TypeEpisodes)), entryVal).WithTTL(time.Hour)
			if err := txn.SetEntry(entry); err != nil {
				return fmt.Errorf("failed to set running man episodes key: %w", err)
			}

			return nil
		}

		valCopy, err := item.ValueCopy(nil)
		if err != nil {
			return fmt.Errorf("failed to copy the value of episodes key: %w", err)
		}

		episodes, err = converter.BytesToInt32Slice(valCopy)
		if err != nil {
			return fmt.Errorf("failed to convert bytes of valCopy to int32: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return episodes, nil
}

func (rme RunningManEpisodes) GenInlineKeyboard(inlineKeyboardType InlineKeyboardType) tg.InlineKeyboardMarkup {
	numOfRowItems := 5
	numOfRows := int(math.Ceil(float64(len(rme.Episodes) / numOfRowItems)))

	inlineKeyboardRows := make([][]tg.InlineKeyboardButton, 0, numOfRows)
	inlineKeyboardRowItems := make([]tg.InlineKeyboardButton, 0, numOfRowItems)

	for _, v := range rme.Episodes {
		btnText := fmt.Sprintf("%d", v)
		btnData := fmt.Sprintf("%s:%d:%d", inlineKeyboardType, rme.Year, v)
		inlineKeyboardRowItems = append(inlineKeyboardRowItems, tg.NewInlineKeyboardButtonData(btnText, btnData))

		if len(inlineKeyboardRowItems) == numOfRowItems {
			inlineKeyboardRows = append(inlineKeyboardRows, tg.NewInlineKeyboardRow(inlineKeyboardRowItems...))
			inlineKeyboardRowItems = inlineKeyboardRowItems[:0]
		}
	}

	if len(inlineKeyboardRowItems) != 0 {
		inlineKeyboardRows = append(inlineKeyboardRows, tg.NewInlineKeyboardRow(inlineKeyboardRowItems...))
	}

	inlineKeyboardRows = append(inlineKeyboardRows, tg.NewInlineKeyboardRow(
		tg.NewInlineKeyboardButtonData("Kembali", fmt.Sprintf("%s:%s", TypeLibraries, "")),
	))

	return tg.NewInlineKeyboardMarkup(inlineKeyboardRows...)
}

func (rme RunningManEpisodes) Process() (tg.Chattable, error) {
	episodes, err := rme.GetRunningManEpisodes()
	if err != nil {
		return nil, err
	}
	rme.Episodes = episodes

	chat := tg.NewEditMessageTextAndMarkup(rme.ChatID, rme.MessageID, EpisodesTextMsg, rme.GenInlineKeyboard(TypeEpisode))
	return chat, nil
}
