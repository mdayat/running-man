package commands

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type BrowseCommand struct {
	ChatID int64
}

func (bc BrowseCommand) Process() (msg tgbotapi.MessageConfig, _ error) {
	msg = tgbotapi.NewMessage(bc.ChatID, "browse command")
	return msg, nil
}
