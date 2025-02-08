package callback

import (
	"context"
	"fmt"
	"strings"

	"github.com/avast/retry-go/v4"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mdayat/running-man/configs/services"
	"github.com/mdayat/running-man/repository"
	"github.com/rs/zerolog/log"
)

var TypeInvoiceProcessor = "invoice_processor"

type invoiceProcessor struct {
	invoiceID string
	isValid   bool
}

func newInvoiceProcessor(invoiceID string) invoiceProcessor {
	return invoiceProcessor{
		invoiceID: invoiceID,
		isValid:   false,
	}
}

func (ip *invoiceProcessor) validateInvoice(ctx context.Context) error {
	invoiceUUID, err := uuid.Parse(ip.invoiceID)
	if err != nil {
		return fmt.Errorf("failed to parse invoice ID string to UUID: %w", err)
	}

	validationResult, err := retry.DoWithData(
		func() (repository.ValidateInvoiceRow, error) {
			return services.Queries.ValidateInvoice(ctx, pgtype.UUID{Bytes: invoiceUUID, Valid: true})
		},
		retry.Attempts(3),
		retry.LastErrorOnly(true),
	)
	if err != nil {
		return err
	}

	if !validationResult.IsExpired && !validationResult.IsUsed {
		ip.isValid = true
	}

	return nil
}

func (ip invoiceProcessor) genInvalidInvoiceMsg(chatID int64) *bot.SendMessageParams {
	text := "Tagihan tidak dapat digunakan karena invalid. Pastikan tagihan belum pernah dipakai dan tidak melebihi 1 jam sejak pembuatan."
	msg := &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	}
	return msg
}

func (ip invoiceProcessor) genValidInvoiceMsg(ctx context.Context, chatID int64) (*bot.SendMessageParams, error) {
	invoiceUUID, err := uuid.Parse(ip.invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse invoice ID string to UUID: %w", err)
	}

	paymentURL, err := retry.DoWithData(
		func() (string, error) {
			return services.Queries.GetPaymentURLByInvoiceID(ctx, pgtype.UUID{Bytes: invoiceUUID, Valid: true})
		},
		retry.Attempts(3),
		retry.LastErrorOnly(true),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment URL by invoice ID: %w", err)
	}

	text := "Yuk segera selesaikan pembayaran dengan menekan tombol \"Proses Pembayaran\"."
	inlineKeyboard := models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Proses Pembayaran", URL: paymentURL},
			},
		},
	}

	return &bot.SendMessageParams{ChatID: chatID, Text: text, ReplyMarkup: inlineKeyboard}, nil
}

func InvoiceProcessorHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	logger := log.Ctx(ctx).With().Int64("user_id", update.CallbackQuery.From.ID).Logger()
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	invoiceID := strings.Split(update.CallbackQuery.Data, ":")[1]
	ip := newInvoiceProcessor(invoiceID)

	logger.Info().Msg("validating invoice...")
	err := ip.validateInvoice(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("failed to validate invoice")
		return
	}
	logger.Info().Msg("successfully validated invoice")

	var msg *bot.SendMessageParams
	if ip.isValid {
		logger.Info().Msg("generating valid invoice message...")
		msg, err = ip.genValidInvoiceMsg(ctx, update.CallbackQuery.Message.Message.Chat.ID)
		if err != nil {
			logger.Error().Err(err).Msg("failed to generate valid invoice message")
			return
		}
		logger.Info().Msg("successfully generated valid invoice message")
	} else {
		msg = ip.genInvalidInvoiceMsg(update.CallbackQuery.Message.Message.Chat.ID)
	}

	logger.Info().Msg("sending invoice processor message...")
	_, err = retry.DoWithData(
		func() (*models.Message, error) {
			return b.SendMessage(ctx, msg)
		},
		retry.Attempts(3),
		retry.LastErrorOnly(true),
	)
	if err != nil {
		logger.Error().Err(err).Msg("failed to send invoice processor message")
		return
	}
	logger.Info().Msg("successfully sent invoice processor message")
}
