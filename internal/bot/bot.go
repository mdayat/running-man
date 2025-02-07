package bot

import (
	"github.com/go-telegram/bot"
	"github.com/mdayat/running-man/internal/callback"
	"github.com/mdayat/running-man/internal/command"
)

func New(botToken string) (*bot.Bot, error) {
	b, err := bot.New(botToken)
	if err != nil {
		return nil, err
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, command.StartHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/browse", bot.MatchTypeExact, command.BrowseHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/support", bot.MatchTypeExact, command.SupportHandler)

	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, callback.TypeLibraries, bot.MatchTypePrefix, callback.LibrariesHandler)

	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, callback.TypeVideoList, bot.MatchTypePrefix, callback.VideoListHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, callback.TypeVideoItem, bot.MatchTypePrefix, callback.VideoItemHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, callback.TypeVideoLink, bot.MatchTypePrefix, callback.VideoLinkHandler)

	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, callback.TypeInvoiceProcessor, bot.MatchTypePrefix, callback.InvoiceProcessorHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, callback.TypeInvoice, bot.MatchTypePrefix, callback.InvoiceHandler)

	return b, nil
}
