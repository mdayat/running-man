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
	TypeVideoCollection    InlineKeyboardType = "video_collection"
	VideoCollectionTextMsg                    = "Pilih episode Running Man dari koleksi video kamu:"
)

type VideoCollection struct {
	ChatID    int64
	UserID    int64
	MessageID int
	Episodes  []int32
}

func (vc VideoCollection) GetEpisodesFromUserVideoCollection() ([]int32, error) {
	var episodes []int32
	err := services.Badger.Update(func(txn *badger.Txn) error {
		videoCollectionKey := fmt.Sprintf("%d:%s", vc.UserID, TypeVideoCollection)
		item, err := txn.Get([]byte(videoCollectionKey))
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return fmt.Errorf("failed to get %s key: %w", videoCollectionKey, err)
		}

		if err != nil && errors.Is(err, badger.ErrKeyNotFound) {
			retryFunc := func() ([]int32, error) {
				return services.Queries.GetEpisodesFromUserVideoCollection(context.TODO(), vc.UserID)
			}

			episodes, err = retry.DoWithData(retryFunc, retry.Attempts(3))
			if err != nil {
				return fmt.Errorf("failed to get episodes from user video collection: %w", err)
			}

			entryVal, err := converter.Int32SliceToBytes(episodes)
			if err != nil {
				return fmt.Errorf("failed to convert int32 of episodes to bytes: %w", err)
			}

			entry := badger.NewEntry([]byte(videoCollectionKey), entryVal).WithTTL(time.Hour)
			if err := txn.SetEntry(entry); err != nil {
				return fmt.Errorf("failed to set %s key: %w", videoCollectionKey, err)
			}

			return nil
		}

		valCopy, err := item.ValueCopy(nil)
		if err != nil {
			return fmt.Errorf("failed to copy the value of %s key: %w", videoCollectionKey, err)
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

func (vc VideoCollection) GenInlineKeyboard(inlineKeyboardType InlineKeyboardType) tg.InlineKeyboardMarkup {
	numOfRowItems := 5
	numOfRows := int(math.Ceil(float64(len(vc.Episodes) / numOfRowItems)))

	inlineKeyboardRows := make([][]tg.InlineKeyboardButton, 0, numOfRows)
	inlineKeyboardRowItems := make([]tg.InlineKeyboardButton, 0, numOfRowItems)

	for _, v := range vc.Episodes {
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

func (vc VideoCollection) Process() (tg.Chattable, error) {
	episodes, err := vc.GetEpisodesFromUserVideoCollection()
	if err != nil {
		return nil, err
	}
	vc.Episodes = episodes

	chat := tg.NewEditMessageTextAndMarkup(vc.ChatID, vc.MessageID, VideoCollectionTextMsg, vc.GenInlineKeyboard("TOOD"))
	return chat, nil
}
