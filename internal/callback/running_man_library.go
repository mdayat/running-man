package callback

import (
	"fmt"
	"math"
	"sort"

	"github.com/avast/retry-go/v4"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var TypeRunningManLibrary InlineKeyboardType = "library"

type Library struct {
	LibraryID string
	Year      int
}

type RunningManLibrary struct {
	ChatID    int64
	MessageID int
	Libraries []Library
}

func (rml RunningManLibrary) GetRunningManLibraries() ([]Library, error) {
	// will be replaced by querying to database
	retryFunc := func() ([]Library, error) {
		result := []Library{
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

func (rml RunningManLibrary) SortLibraries() {
	sort.Slice(rml.Libraries, func(i, j int) bool {
		return rml.Libraries[i].Year < rml.Libraries[j].Year
	})
}

func (rml RunningManLibrary) GenInlineKeyboard(inlineKeyboardType InlineKeyboardType) tg.InlineKeyboardMarkup {
	numOfRowItems := 3
	numOfRows := int(math.Ceil(float64(len(rml.Libraries) / numOfRowItems)))

	inlineKeyboardRows := make([][]tg.InlineKeyboardButton, 0, numOfRows)
	inlineKeyboardRowItems := make([]tg.InlineKeyboardButton, 0, numOfRowItems)

	for _, v := range rml.Libraries {
		btnText := fmt.Sprintf("%d", v.Year)
		btnData := fmt.Sprintf("%s:%s", inlineKeyboardType, v.LibraryID)
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

func (rml RunningManLibrary) Process() (tg.Chattable, error) {
	libraries, err := rml.GetRunningManLibraries()
	if err != nil {
		return nil, fmt.Errorf("failed to get running man libraries: %w", err)
	}

	rml.Libraries = libraries
	rml.SortLibraries()

	chat := tg.NewEditMessageTextAndMarkup(
		rml.ChatID,
		rml.MessageID,
		"Pilih tahun Running Man:",
		rml.GenInlineKeyboard(TypeRunningManEpisode),
	)

	return chat, nil
}
