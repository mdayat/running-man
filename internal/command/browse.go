package command

import (
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mdayat/running-man/internal/callback"
)

type Browse struct {
	ChatID int64
}

func (b Browse) Process() (tg.Chattable, error) {
	var rml callback.RunningManLibraries
	years, err := rml.GetRunningManYears()
	if err != nil {
		return nil, err
	}
	rml.Years = years

	chat := tg.NewMessage(b.ChatID, callback.LibrariesTextMsg)
	chat.ReplyMarkup = rml.GenInlineKeyboard(callback.TypeEpisodes)
	return chat, nil
}
