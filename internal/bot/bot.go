package bot

import (
	"github.com/go-telegram/bot"
	"github.com/mdayat/running-man/internal/command"
)

func New(botToken string) (*bot.Bot, error) {
	b, err := bot.New(botToken)
	if err != nil {
		return nil, err
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, command.DefaultHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypeExact, command.DefaultHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/browse", bot.MatchTypeExact, command.BrowseHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/collection", bot.MatchTypeExact, command.CollectionHandler)

	return b, nil
}
