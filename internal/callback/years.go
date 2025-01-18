package callback

import (
	"fmt"
	"math"
	"sort"

	"github.com/avast/retry-go/v4"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	TypeYears    InlineKeyboardType = "years"
	YearsTextMsg                    = "Pilih tahun Running Man:"
)

type RunningManYears struct {
	ChatID    int64
	MessageID int
	Years     []int
}

func (rml RunningManYears) GetRunningManYears() ([]int, error) {
	// will be replaced by querying to database
	retryFunc := func() ([]int, error) {
		result := []int{2020, 2018, 2015, 2017, 2016, 2011}
		return result, nil
	}

	result, err := retry.DoWithData(retryFunc, retry.Attempts(3))
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (rml RunningManYears) SortYears() {
	sort.Slice(rml.Years, func(i, j int) bool {
		return rml.Years[i] < rml.Years[j]
	})
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
		return nil, fmt.Errorf("failed to get running man years: %w", err)
	}

	rml.Years = years
	rml.SortYears()

	chat := tg.NewEditMessageTextAndMarkup(rml.ChatID, rml.MessageID, YearsTextMsg, rml.GenInlineKeyboard(TypeEpisodes))
	return chat, nil
}
