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
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/mdayat/running-man/configs/services"
	"github.com/mdayat/running-man/internal/converter"
	"github.com/rs/zerolog/log"
)

var (
	TypeVideos    = "videos"
	VideosTextMsg = "Pilih episode Running Man:"
)

type RunningManVideos struct {
	Year      int32
	ChatID    int64
	MessageID int
	Episodes  []int32
}

func (rmv RunningManVideos) GetRunningManEpisodes(ctx context.Context) ([]int32, error) {
	var episodes []int32
	err := services.Badger.Update(func(txn *badger.Txn) error {
		runningManVideosKey := fmt.Sprintf("%d:%s", rmv.Year, TypeVideos)
		item, err := txn.Get([]byte(runningManVideosKey))
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return fmt.Errorf("failed to get %s key: %w", runningManVideosKey, err)
		}

		if err != nil && errors.Is(err, badger.ErrKeyNotFound) {
			retryFunc := func() ([]int32, error) {
				return services.Queries.GetRunningManEpisodesByYear(ctx, rmv.Year)
			}

			episodes, err = retry.DoWithData(retryFunc, retry.Attempts(3), retry.LastErrorOnly(true))
			if err != nil {
				return fmt.Errorf("failed to get running man episodes: %w", err)
			}

			entryVal, err := converter.Int32SliceToBytes(episodes)
			if err != nil {
				return fmt.Errorf("failed to convert int32 of episodes to bytes: %w", err)
			}

			entry := badger.NewEntry([]byte(runningManVideosKey), entryVal).WithTTL(time.Hour)
			if err := txn.SetEntry(entry); err != nil {
				return fmt.Errorf("failed to set %s key: %w", runningManVideosKey, err)
			}

			return nil
		}

		valCopy, err := item.ValueCopy(nil)
		if err != nil {
			return fmt.Errorf("failed to copy the value of %s key: %w", runningManVideosKey, err)
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

func (rmv RunningManVideos) GenInlineKeyboard(inlineKeyboardType string) tg.InlineKeyboardMarkup {
	numOfRowItems := 5
	numOfRows := int(math.Ceil(float64(len(rmv.Episodes) / numOfRowItems)))

	inlineKeyboardRows := make([][]tg.InlineKeyboardButton, 0, numOfRows)
	inlineKeyboardRowItems := make([]tg.InlineKeyboardButton, 0, numOfRowItems)

	for _, v := range rmv.Episodes {
		btnText := fmt.Sprintf("%d", v)
		btnData := fmt.Sprintf("%s:%d,%d", inlineKeyboardType, rmv.Year, v)
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

func VideosHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
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

	rmv := RunningManVideos{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Year:      int32(year),
	}

	episodes, err := rmv.GetRunningManEpisodes(ctx)
	if err != nil {
		logger.Err(err).Send()
		return
	}
	rmv.Episodes = episodes

	_, err = retry.DoWithData(
		func() (*models.Message, error) {
			return b.EditMessageText(ctx, &bot.EditMessageTextParams{
				ChatID:      rmv.ChatID,
				MessageID:   rmv.MessageID,
				Text:        VideosTextMsg,
				ReplyMarkup: rmv.GenInlineKeyboard(TypeVideo),
			})
		},
		retry.Attempts(3),
		retry.LastErrorOnly(true),
	)

	if err != nil {
		logger.Err(err).Msgf("failed to send %s callback edit message", TypeVideos)
		return
	}
}
