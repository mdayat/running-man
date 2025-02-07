package callback

import (
	"context"
	"fmt"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mdayat/running-man/configs/services"
	"github.com/mdayat/running-man/internal/tripay"
	"github.com/mdayat/running-man/repository"
	"github.com/rs/zerolog/log"
)

var TypeInvoice = "invoice"

func handleHasValidInvoice(ctx context.Context, b *bot.Bot, chatID int64) error {
	text := "Tidak dapat membuat tagihan karena kamu telah memiliki tagihan yang masih valid. Gunakan tagihan tersebut untuk melakukan pembayaran."
	_, err := retry.DoWithData(
		func() (*models.Message, error) {
			return b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   text,
			})
		},
		retry.Attempts(3),
		retry.LastErrorOnly(true),
	)

	return err
}

func InvoiceHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	logger := log.Ctx(ctx).With().Int64("user_id", update.CallbackQuery.From.ID).Logger()
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	logger.Info().Msg("checking if the user has valid invoice...")
	hasValidInvoice, err := retry.DoWithData(
		func() (bool, error) {
			return services.Queries.HasValidInvoice(ctx, update.CallbackQuery.From.ID)
		},
		retry.Attempts(3),
		retry.LastErrorOnly(true),
	)

	if err != nil {
		logger.Error().Err(err).Msg("failed to check if the user has valid invoice")
		return
	}
	logger.Info().Msg("successfully checked if the user has valid invoice")

	if hasValidInvoice {
		logger.Info().Msg("the user has valid invoice. sending a telegram message to the user...")
		if err := handleHasValidInvoice(ctx, b, update.CallbackQuery.From.ID); err != nil {
			logger.Error().Err(err).Msg("failed to send a telegram message to the user")
			return
		}
		logger.Info().Msg("successfully sent a telegram message to the user")
		return
	}
	logger.Info().Msg("the user hasn't valid invoice")

	merchantRef := uuid.New()
	merchantRefString := merchantRef.String()
	subscriptionPrice := 1000

	orderedItems := []tripay.OrderedItem{
		{
			Name:     "Running Man Subscription",
			Price:    subscriptionPrice,
			Quantity: 1,
		},
	}

	params := tripay.NewTransactionBodyParams{
		MerchantRef:   merchantRefString,
		CustomerName:  update.CallbackQuery.From.FirstName,
		CustomerEmail: "odemimasa@gmail.com",
		TotalAmount:   subscriptionPrice,
		OrderedItems:  orderedItems,
	}

	logger.Info().Msg("requesting tripay transaction...")
	response, err := tripay.RequestTransaction(ctx, tripay.NewTransactionBody(params))
	if err != nil {
		logger.Error().Err(err).Msg("failed to request tripay transaction")
		return
	}
	logger.Info().Msg("successfully requested tripay transaction")

	text := "Tagihan berhasil dibuat! Tagihan akan kedaluwarsa setelah satu jam.\n\nKlik tombol \"Proses Tagihan\" untuk melanjutkan ke proses pembayaran."
	inlineKeyboard := models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Proses Tagihan", CallbackData: fmt.Sprintf("%s:%s", TypeInvoiceProcessor, merchantRefString)},
			},
		},
	}

	logger.Info().Msg("creating and sending invoice...")
	retryFunc := func() (err error) {
		var tx pgx.Tx
		tx, err = services.DB.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction to create and send invoice: %w", err)
		}

		defer func() {
			if err == nil {
				err = tx.Commit(ctx)
			}

			if err != nil {
				tx.Rollback(ctx)
			}
		}()

		qtx := services.Queries.WithTx(tx)
		err = qtx.CreateInvoice(ctx, repository.CreateInvoiceParams{
			ID:          pgtype.UUID{Bytes: merchantRef, Valid: true},
			UserID:      update.CallbackQuery.From.ID,
			RefID:       response.Reference,
			QrUrl:       response.QrURL,
			TotalAmount: int32(response.Amount),
			ExpiredAt:   pgtype.Timestamptz{Time: time.Unix(int64(response.ExpiredTime), 0), Valid: true},
		})

		if err != nil {
			return fmt.Errorf("failed to create invoice: %w", err)
		}

		msg := &bot.SendMessageParams{ChatID: update.CallbackQuery.From.ID, Text: text, ReplyMarkup: inlineKeyboard}
		if _, err = b.SendMessage(ctx, msg); err != nil {
			return fmt.Errorf("failed to send invoice: %w", err)
		}

		return nil
	}

	if err := retry.Do(retryFunc, retry.Attempts(3), retry.LastErrorOnly(true)); err != nil {
		logger.Error().Err(err).Msg("failed to create and send invoice")
		return
	}
	logger.Info().Msg("successfully created and sent invoice")
}
