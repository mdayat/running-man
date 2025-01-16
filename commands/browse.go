package commands

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type Browse struct {
	ChatID int64
}

func (b Browse) Process() (chat tgbotapi.Chattable, _ error) {
	chat = tgbotapi.NewMessage(b.ChatID, "browse command")
	return chat, nil
}
