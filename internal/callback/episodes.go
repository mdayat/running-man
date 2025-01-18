package callback

import (
	"fmt"
	"math"
	"sort"

	"github.com/avast/retry-go/v4"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	TypeEpisodes    InlineKeyboardType = "episodes"
	EpisodesTextMsg                    = "Pilih episode Running Man:"
)

type RunningManEpisodes struct {
	Year      int
	ChatID    int64
	MessageID int
	Episodes  []int
}

func (rme RunningManEpisodes) GetRunningManEpisodes() ([]int, error) {
	// will be replaced by querying to database
	retryFunc := func() ([]int, error) {
		result := []int{420, 418, 415, 417, 416, 411}
		return result, nil
	}

	result, err := retry.DoWithData(retryFunc, retry.Attempts(3))
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (rme RunningManEpisodes) SortEpisodes() {
	sort.Slice(rme.Episodes, func(i, j int) bool {
		return rme.Episodes[i] < rme.Episodes[j]
	})
}

func (rme RunningManEpisodes) GenInlineKeyboard(inlineKeyboardType InlineKeyboardType) tg.InlineKeyboardMarkup {
	numOfRowItems := 5
	numOfRows := int(math.Ceil(float64(len(rme.Episodes) / numOfRowItems)))

	inlineKeyboardRows := make([][]tg.InlineKeyboardButton, 0, numOfRows)
	inlineKeyboardRowItems := make([]tg.InlineKeyboardButton, 0, numOfRowItems)

	for _, v := range rme.Episodes {
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

	inlineKeyboardRows = append(inlineKeyboardRows, tg.NewInlineKeyboardRow(
		tg.NewInlineKeyboardButtonData("Kembali", fmt.Sprintf("%s:%s", TypeYears, "")),
	))

	return tg.NewInlineKeyboardMarkup(inlineKeyboardRows...)
}

func (rme RunningManEpisodes) Process() (tg.Chattable, error) {
	episodes, err := rme.GetRunningManEpisodes()
	if err != nil {
		return nil, fmt.Errorf("failed to get running man episodes: %w", err)
	}

	rme.Episodes = episodes
	rme.SortEpisodes()

	chat := tg.NewEditMessageTextAndMarkup(rme.ChatID, rme.MessageID, EpisodesTextMsg, rme.GenInlineKeyboard("TODO"))
	return chat, nil
}
