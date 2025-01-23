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
	TypeVideoCollection    = "video_collection"
	VideoCollectionTextMsg = "Pilih episode Running Man dari koleksi video kamu:"
)

type VideoCollection struct {
	ChatID    int64
	UserID    int64
	MessageID int
	Episodes  []int32
}

func (vc VideoCollection) GetEpisodesFromUserVideoCollection(ctx context.Context) ([]int32, error) {
	var episodes []int32
	err := services.Badger.Update(func(txn *badger.Txn) error {
		videoCollectionKey := fmt.Sprintf("%d:%s", vc.UserID, TypeVideoCollection)
		item, err := txn.Get([]byte(videoCollectionKey))
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return fmt.Errorf("failed to get %s key: %w", videoCollectionKey, err)
		}

		if err != nil && errors.Is(err, badger.ErrKeyNotFound) {
			retryFunc := func() ([]int32, error) {
				return services.Queries.GetEpisodesFromUserVideoCollection(ctx, vc.UserID)
			}

			episodes, err = retry.DoWithData(retryFunc, retry.Attempts(3), retry.LastErrorOnly(true))
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

func (vc VideoCollection) GenInlineKeyboard(inlineKeyboardType string) models.InlineKeyboardMarkup {
	numOfRowItems := 5
	numOfRows := int(math.Ceil(float64(len(vc.Episodes)) / float64(numOfRowItems)))

	inlineKeyboardRows := make([][]models.InlineKeyboardButton, 0, numOfRows)
	inlineKeyboardRowItems := make([]models.InlineKeyboardButton, 0, numOfRowItems)

	for _, v := range vc.Episodes {
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

func VideoCollectionHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	logger := log.Ctx(ctx).With().Logger()
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	vc := VideoCollection{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		UserID:    update.CallbackQuery.From.ID,
	}

	episodes, err := vc.GetEpisodesFromUserVideoCollection(ctx)
	if err != nil {
		logger.Err(err).Send()
		return
	}
	vc.Episodes = episodes

	_, err = retry.DoWithData(
		func() (*models.Message, error) {
			return b.EditMessageText(ctx, &bot.EditMessageTextParams{
				ChatID:      vc.ChatID,
				MessageID:   vc.MessageID,
				Text:        VideoCollectionTextMsg,
				ReplyMarkup: vc.GenInlineKeyboard(TypeVideoCollectionItem),
			})
		},
		retry.Attempts(3),
		retry.LastErrorOnly(true),
	)

	if err != nil {
		logger.Err(err).Msgf("failed to send %s callback edit message", TypeVideoCollection)
		return
	}
}
