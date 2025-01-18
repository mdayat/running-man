package command

import (
	"fmt"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mdayat/running-man/internal/callback"
)

type Browse struct {
	ChatID int64
}

func (b Browse) Process() (tg.Chattable, error) {
	var rml callback.RunningManYears
	years, err := rml.GetRunningManYears()
	if err != nil {
		return nil, fmt.Errorf("failed to get running man years: %w", err)
	}

	rml.Years = years
	rml.SortYears()

	chat := tg.NewMessage(b.ChatID, callback.RunningManYearTextMsg)
	chat.ReplyMarkup = rml.GenInlineKeyboard(callback.TypeRunningManEpisode)
	return chat, nil
}
