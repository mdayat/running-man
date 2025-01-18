package callback

import (
	"fmt"
	"math"
	"sort"

	"github.com/avast/retry-go/v4"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var TypeRunningManEpisode InlineKeyboardType = "episode"

type Episode struct {
	MovieID string
	Episode int
}

type RunningManEpisode struct {
	LibraryID string
	ChatID    int64
	MessageID int
	Episodes  []Episode
}

func (rme RunningManEpisode) GetRunningManEpisodes() ([]Episode, error) {
	// will be replaced by querying to database
	retryFunc := func() ([]Episode, error) {
		result := []Episode{
			{MovieID: "123132", Episode: 420},
			{MovieID: "123132", Episode: 418},
			{MovieID: "123132", Episode: 415},
			{MovieID: "123132", Episode: 417},
			{MovieID: "123132", Episode: 416},
			{MovieID: "123132", Episode: 411},
		}

		return result, nil
	}

	result, err := retry.DoWithData(retryFunc, retry.Attempts(3))
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (rme RunningManEpisode) SortEpisodes() {
	sort.Slice(rme.Episodes, func(i, j int) bool {
		return rme.Episodes[i].Episode < rme.Episodes[j].Episode
	})
}

func (rme RunningManEpisode) GenInlineKeyboard(inlineKeyboardType InlineKeyboardType) tg.InlineKeyboardMarkup {
	numOfRowItems := 5
	numOfRows := int(math.Ceil(float64(len(rme.Episodes) / numOfRowItems)))

	inlineKeyboardRows := make([][]tg.InlineKeyboardButton, 0, numOfRows)
	inlineKeyboardRowItems := make([]tg.InlineKeyboardButton, 0, numOfRowItems)

	for _, v := range rme.Episodes {
		btnText := fmt.Sprintf("%d", v.Episode)
		btnData := fmt.Sprintf("%s:%s", inlineKeyboardType, v.MovieID)
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
		tg.NewInlineKeyboardButtonData("Kembali", fmt.Sprintf("%s:%s", TypeRunningManLibrary, "")),
	))

	return tg.NewInlineKeyboardMarkup(inlineKeyboardRows...)
}

func (rme RunningManEpisode) Process() (tg.Chattable, error) {
	episodes, err := rme.GetRunningManEpisodes()
	if err != nil {
		return nil, fmt.Errorf("failed to get running man episodes: %w", err)
	}

	rme.Episodes = episodes
	rme.SortEpisodes()

	chat := tg.NewEditMessageTextAndMarkup(
		rme.ChatID,
		rme.MessageID,
		"Pilih episode Running Man:",
		rme.GenInlineKeyboard("TODO"),
	)

	return chat, nil
}
