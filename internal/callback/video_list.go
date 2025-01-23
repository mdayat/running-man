package callback

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
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
	TypeVideoList    = "video_list"
	VideoListTextMsg = "Pilih episode Running Man:"
)

type VideoList struct {
	Year      int32
	ChatID    int64
	MessageID int
	Episodes  []int32
}

func (vl VideoList) GetRunningManEpisodes(ctx context.Context) ([]int32, error) {
	var episodes []int32
	err := services.Badger.Update(func(txn *badger.Txn) error {
		videoListKey := fmt.Sprintf("%d:%s", vl.Year, TypeVideoList)
		item, err := txn.Get([]byte(videoListKey))
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return fmt.Errorf("failed to get %s key: %w", videoListKey, err)
		}

		if err != nil && errors.Is(err, badger.ErrKeyNotFound) {
			retryFunc := func() ([]int32, error) {
				return services.Queries.GetRunningManEpisodesByYear(ctx, vl.Year)
			}

			episodes, err = retry.DoWithData(retryFunc, retry.Attempts(3), retry.LastErrorOnly(true))
			if err != nil {
				return fmt.Errorf("failed to get running man episodes: %w", err)
			}

			entryVal, err := converter.Int32SliceToBytes(episodes)
			if err != nil {
				return fmt.Errorf("failed to convert int32 of episodes to bytes: %w", err)
			}

			entry := badger.NewEntry([]byte(videoListKey), entryVal).WithTTL(time.Hour)
			if err := txn.SetEntry(entry); err != nil {
				return fmt.Errorf("failed to set %s key: %w", videoListKey, err)
			}

			return nil
		}

		valCopy, err := item.ValueCopy(nil)
		if err != nil {
			return fmt.Errorf("failed to copy the value of %s key: %w", videoListKey, err)
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

func (vl VideoList) GenInlineKeyboard(inlineKeyboardType string) models.InlineKeyboardMarkup {
	numOfRowItems := 5
	numOfRows := int(math.Ceil(float64(len(vl.Episodes)) / float64(numOfRowItems)))

	inlineKeyboardRows := make([][]models.InlineKeyboardButton, 0, numOfRows)
	inlineKeyboardRowItems := make([]models.InlineKeyboardButton, 0, numOfRowItems)

	for _, v := range vl.Episodes {
		btnText := fmt.Sprintf("%d", v)
		btnData := fmt.Sprintf("%s:%d,%d", inlineKeyboardType, vl.Year, v)
		inlineKeyboardRowItems = append(inlineKeyboardRowItems, models.InlineKeyboardButton{Text: btnText, CallbackData: btnData})

		if len(inlineKeyboardRowItems) == numOfRowItems {
			inlineKeyboardRows = append(inlineKeyboardRows, inlineKeyboardRowItems)
			inlineKeyboardRowItems = make([]models.InlineKeyboardButton, 0, numOfRowItems)
		}
	}

	if len(inlineKeyboardRowItems) != 0 {
		inlineKeyboardRows = append(inlineKeyboardRows, inlineKeyboardRowItems)
	}

	inlineKeyboardRows = append(inlineKeyboardRows, []models.InlineKeyboardButton{
		{Text: "Kembali", CallbackData: fmt.Sprintf("%s:%s", TypeLibraries, "")},
	})

	return models.InlineKeyboardMarkup{InlineKeyboard: inlineKeyboardRows}
}

func VideoListHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	logger := log.Ctx(ctx).With().Logger()
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	year, err := strconv.Atoi(strings.Split(update.CallbackQuery.Data, ":")[1])
	if err != nil {
		logger.Err(err).Msg("failed to convert year string to int")
		return
	}

	vl := VideoList{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Year:      int32(year),
	}

	episodes, err := vl.GetRunningManEpisodes(ctx)
	if err != nil {
		logger.Err(err).Send()
		return
	}
	vl.Episodes = episodes

	_, err = retry.DoWithData(
		func() (*models.Message, error) {
			return b.EditMessageText(ctx, &bot.EditMessageTextParams{
				ChatID:      vl.ChatID,
				MessageID:   vl.MessageID,
				Text:        VideoListTextMsg,
				ReplyMarkup: vl.GenInlineKeyboard(TypeVideoItem),
			})
		},
		retry.Attempts(3),
		retry.LastErrorOnly(true),
	)

	if err != nil {
		logger.Err(err).Msgf("failed to send %s callback edit message", TypeVideoList)
		return
	}
}
