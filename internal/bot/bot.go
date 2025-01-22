package bot

import (
	"context"
	"encoding/json"

	"github.com/avast/retry-go/v4"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/mdayat/running-man/internal/callback"
	"github.com/mdayat/running-man/internal/command"
	"github.com/mdayat/running-man/internal/payment"
	"github.com/rs/zerolog/log"
)

func New(botToken string) (*bot.Bot, error) {
	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
	}

	b, err := bot.New(botToken, opts...)
	if err != nil {
		return nil, err
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, command.DefaultHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypeExact, command.DefaultHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/browse", bot.MatchTypeExact, command.BrowseHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/collection", bot.MatchTypeExact, command.CollectionHandler)

	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, callback.TypeLibraries, bot.MatchTypePrefix, callback.LibrariesHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, callback.TypeVideoList, bot.MatchTypePrefix, callback.VideoListHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, callback.TypeVideoItem, bot.MatchTypePrefix, callback.VideoItemHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, callback.TypeInvoice, bot.MatchTypePrefix, callback.InvoiceHandler)

	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, callback.TypeVideoCollectionItem, bot.MatchTypePrefix, callback.VideoCollectionItemHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, callback.TypeVideoCollection, bot.MatchTypePrefix, callback.VideoCollectionHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, callback.TypeVideoLink, bot.MatchTypePrefix, callback.VideoLinkHandler)

	return b, nil
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	logger := log.Ctx(ctx).With().Logger()

	if update.PreCheckoutQuery != nil {
		var payload payment.InvoicePayload
		if err := json.Unmarshal([]byte(update.PreCheckoutQuery.InvoicePayload), &payload); err != nil {
			logger.Err(err).Msg("failed to unmarshal invoice payload")
			return
		}

		isUserHasVideo, err := payment.CheckVideoOwnership(ctx, payment.CheckVideoOwnershipParams{
			UserID:  payload.UserID,
			Episode: payload.Episode,
		})

		if err != nil {
			logger.Err(err).Msg("failed to check video ownership")
			return
		}

		isInvoiceExpired, err := payment.CheckExpiredInvoice(ctx, payload.ID)
		if err != nil {
			logger.Err(err).Msg("failed to check expired invoice")
			return
		}

		if isUserHasVideo || isInvoiceExpired {
			var text string
			if isUserHasVideo {
				text = payment.MakeVideoOwnershipText(payload.Episode)
			} else {
				text = payment.MakeExpiredInvoiceText(payload.Episode)
			}

			_, err = retry.DoWithData(
				func() (bool, error) {
					return b.AnswerPreCheckoutQuery(ctx, &bot.AnswerPreCheckoutQueryParams{
						PreCheckoutQueryID: update.PreCheckoutQuery.ID,
						OK:                 false,
						ErrorMessage:       text,
					})
				},
				retry.Attempts(3),
				retry.LastErrorOnly(true),
			)

			if err != nil {
				logger.Err(err).Msg("failed to answer false pre-checkout query")
				return
			}
			return
		}

		_, err = retry.DoWithData(
			func() (bool, error) {
				return b.AnswerPreCheckoutQuery(ctx, &bot.AnswerPreCheckoutQueryParams{
					PreCheckoutQueryID: update.PreCheckoutQuery.ID,
					OK:                 true,
					ErrorMessage:       "",
				})
			},
			retry.Attempts(3),
			retry.LastErrorOnly(true),
		)

		if err != nil {
			logger.Err(err).Msg("failed to answer true pre-checkout query")
			return
		}
		return
	}

	if update.Message != nil && update.Message.SuccessfulPayment != nil {
		var payload payment.InvoicePayload
		if err := json.Unmarshal([]byte(update.Message.SuccessfulPayment.InvoicePayload), &payload); err != nil {
			logger.Err(err).Msg("failed to unmarshal invoice payload")
			return
		}

		err := payment.InsertAndSendSuccessfulPayment(ctx, b, payment.InsertAndSendSuccessfulPaymentParams{
			Payload:   payload,
			PaymentID: update.Message.SuccessfulPayment.TelegramPaymentChargeID,
		})

		if err != nil {
			logger.Err(err).Send()
			return
		}
		return
	}
}
