package command

import (
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mdayat/running-man/internal/callback"
)

type Collection struct {
	ChatID int64
	UserID int64
}

func (c Collection) Process() (tg.Chattable, error) {
	vc := callback.VideoCollection{
		ChatID: c.ChatID,
		UserID: c.UserID,
	}

	episodes, err := vc.GetEpisodesFromUserVideoCollection()
	if err != nil {
		return nil, err
	}
	vc.Episodes = episodes

	chat := tg.NewMessage(c.ChatID, callback.VideoCollectionTextMsg)
	chat.ReplyMarkup = vc.GenInlineKeyboard("TODO")
	return chat, nil
}
