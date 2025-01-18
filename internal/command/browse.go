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
	rml := callback.RunningManLibrary{}
	libraries, err := rml.GetRunningManLibraries()
	if err != nil {
		return nil, fmt.Errorf("failed to get running man libraries: %w", err)
	}

	rml.Libraries = libraries
	rml.SortLibraries()

	chat := tg.NewMessage(b.ChatID, "Pilih tahun Running Man:")
	chat.ReplyMarkup = rml.GenInlineKeyboard(callback.TypeRunningManEpisode)
	return chat, nil
}
