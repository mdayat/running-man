package command

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mdayat/running-man/internal/callback"
)

type Browse struct {
	ChatID int64
}

func (b Browse) Process() (tgbotapi.Chattable, error) {
	var rmLibraries callback.RunningManLibraries
	rmLibraries, err := rmLibraries.GetRunningManLibraries()
	if err != nil {
		return nil, fmt.Errorf("failed to get running man libraries: %w", err)
	}
	rmLibraries.Sort()

	chat := tgbotapi.NewMessage(b.ChatID, "Pilih tahun Running Man dari daftar di bawah ini:")
	chat.ReplyMarkup = rmLibraries.GenInlineKeyboard(callback.TypeRunningManLibrary)
	return chat, nil
}
