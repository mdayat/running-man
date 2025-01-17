package callback

import (
	"fmt"
	"math"
	"sort"

	"github.com/avast/retry-go/v4"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var TypeRunningManLibrary InlineKeyboardType = "library"

type RunningManLibrary struct {
	LibraryID string
	Year      int
}

type RunningManLibraries []RunningManLibrary

func (rml RunningManLibraries) GetRunningManLibraries() (RunningManLibraries, error) {
	// will be replaced by querying to database
	retryFunc := func() (RunningManLibraries, error) {
		result := RunningManLibraries{
			{LibraryID: "123132", Year: 2020},
			{LibraryID: "123132", Year: 2018},
			{LibraryID: "123132", Year: 2015},
			{LibraryID: "123132", Year: 2017},
			{LibraryID: "123132", Year: 2016},
			{LibraryID: "123132", Year: 2011},
		}
		return result, nil
	}

	result, err := retry.DoWithData(retryFunc, retry.Attempts(3))
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (rml RunningManLibraries) Sort() {
	sort.Slice(rml, func(i, j int) bool {
		return rml[i].Year < rml[j].Year
	})
}

func (rml RunningManLibraries) GenInlineKeyboard(inlineKeyboardType InlineKeyboardType) tgbotapi.InlineKeyboardMarkup {
	numOfRowItems := 3
	numOfRows := int(math.Ceil(float64(len(rml) / numOfRowItems)))

	inlineKeyboardRows := make([][]tgbotapi.InlineKeyboardButton, 0, numOfRows)
	inlineKeyboardRowItems := make([]tgbotapi.InlineKeyboardButton, 0, numOfRowItems)

	for i := 0; i < len(rml); i++ {
		btnText := fmt.Sprintf("%d", rml[i].Year)
		btnData := fmt.Sprintf("%s:%s", inlineKeyboardType, rml[i].LibraryID)
		inlineKeyboardRowItems = append(inlineKeyboardRowItems, tgbotapi.NewInlineKeyboardButtonData(
			btnText, btnData,
		))

		if len(inlineKeyboardRowItems) == 3 {
			inlineKeyboardRows = append(inlineKeyboardRows, tgbotapi.NewInlineKeyboardRow(inlineKeyboardRowItems...))
			inlineKeyboardRowItems = inlineKeyboardRowItems[:0]
		}
	}

	if len(inlineKeyboardRowItems) != 0 {
		inlineKeyboardRows = append(inlineKeyboardRows, tgbotapi.NewInlineKeyboardRow(inlineKeyboardRowItems...))
	}

	return tgbotapi.NewInlineKeyboardMarkup(inlineKeyboardRows...)
}
