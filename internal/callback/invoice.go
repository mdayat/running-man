package callback

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mdayat/running-man/configs/services"
	"github.com/mdayat/running-man/internal/payment"
	"github.com/mdayat/running-man/repository"
	"github.com/rs/zerolog/log"
)

var TypeInvoice = "invoice"

type Invoice struct {
	MessageID int
	Year      int32
	Payload   payment.InvoicePayload
}

func (i Invoice) GenInvoiceMsg(ctx context.Context) (*bot.SendInvoiceParams, error) {
	price, err := retry.DoWithData(
		func() (int32, error) {
			return services.Queries.GetRunningManVideoPrice(ctx, i.Payload.Episode)
		},
		retry.Attempts(3),
		retry.LastErrorOnly(true),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get running man video price: %w", err)
	}

	tax := int(math.Ceil(float64(price) * 0.11))
	priceAfterTax := int(price) + tax

	payloadJSON, err := json.Marshal(i.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal invoice payload: %w", err)
	}

	title := fmt.Sprintf("Running Man Episode %d", i.Payload.Episode)
	description := fmt.Sprintf(
		"Pembelian Running Man episode %d (%d), harga sudah termasuk pajak. Segera lakukan pembayaran karena tagihan akan invalid setelah 1 jam.",
		i.Payload.Episode,
		i.Year,
	)

	invoiceMsg := bot.SendInvoiceParams{
		ChatID:      i.Payload.ChatID,
		Title:       title,
		Description: description,
		Payload:     string(payloadJSON),
		Currency:    "XTR",
		Prices: []models.LabeledPrice{
			{Label: title, Amount: priceAfterTax},
		},
	}

	return &invoiceMsg, nil
}

func InvoiceHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	logger := log.Ctx(ctx).With().Logger()
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	splittedData := strings.Split(update.CallbackQuery.Data, ":")[1]
	year, err := strconv.Atoi(strings.Split(splittedData, ",")[0])
	if err != nil {
		logger.Err(err).Msg("failed to convert year string to int")
		return
	}

	episode, err := strconv.Atoi(strings.Split(splittedData, ",")[1])
	if err != nil {
		logger.Err(err).Msg("failed to convert episode string to int")
		return
	}

	invoiceUUID := uuid.New()
	i := Invoice{
		MessageID: update.CallbackQuery.Message.Message.ID,
		Year:      int32(year),
		Payload: payment.InvoicePayload{
			ID:      invoiceUUID.String(),
			ChatID:  update.CallbackQuery.Message.Message.Chat.ID,
			UserID:  update.CallbackQuery.From.ID,
			Episode: int32(episode),
		},
	}

	isUserHasVideo, err := payment.CheckVideoOwnership(ctx, payment.CheckVideoOwnershipParams{
		UserID:  i.Payload.UserID,
		Episode: i.Payload.Episode,
	})

	if err != nil {
		logger.Err(err).Msg("failed to check video ownership")
		return
	}

	isInvoiceUnexpired, err := payment.CheckInvoiceExpiration(ctx, payment.CheckInvoiceExpirationParams{
		UserID:  i.Payload.UserID,
		Episode: i.Payload.Episode,
	})

	if err != nil {
		logger.Err(err).Msg("failed to check invoice expiration")
		return
	}

	if isUserHasVideo || isInvoiceUnexpired {
		var msg *bot.SendMessageParams
		if isUserHasVideo {
			msg = &bot.SendMessageParams{
				ChatID: i.Payload.ChatID,
				Text:   payment.MakeVideoOwnershipText(i.Payload.Episode),
			}
		} else {
			msg = &bot.SendMessageParams{
				ChatID: i.Payload.ChatID,
				Text:   payment.MakeUnexpiredInvoiceText(i.Payload.Episode),
			}
		}

		_, err = retry.DoWithData(
			func() (*models.Message, error) {
				return b.SendMessage(ctx, msg)
			},
			retry.Attempts(3),
			retry.LastErrorOnly(true),
		)

		if err != nil {
			logger.Err(err).Msgf("failed to send %s callback message", TypeInvoice)
			return
		}
		return
	}

	invoiceMsg, err := i.GenInvoiceMsg(ctx)
	if err != nil {
		logger.Err(err).Msg("failed to generate invoice message")
		return
	}

	retryFunc := func() (err error) {
		var tx pgx.Tx
		tx, err = services.DB.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
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
			ID:                     pgtype.UUID{Bytes: invoiceUUID, Valid: true},
			UserID:                 i.Payload.UserID,
			RunningManVideoEpisode: i.Payload.Episode,
			Amount:                 int32(invoiceMsg.Prices[0].Amount),
			ExpiredAt:              pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
		})

		if err != nil {
			return fmt.Errorf("failed to insert invoice: %w", err)
		}

		if _, err = b.SendInvoice(ctx, invoiceMsg); err != nil {
			return fmt.Errorf("failed to send %s callback invoice message: %w", TypeInvoice, err)
		}

		return nil
	}

	if err := retry.Do(retryFunc, retry.Attempts(3), retry.LastErrorOnly(true)); err != nil {
		logger.Err(err).Send()
		return
	}
}
