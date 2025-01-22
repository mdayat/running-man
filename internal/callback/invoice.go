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
	"github.com/mdayat/running-man/repository"
	"github.com/rs/zerolog/log"
)

var TypeInvoice = "invoice"

type InvoicePayload struct {
	ID      string `json:"id"`
	UserID  int64  `json:"user_id"`
	Episode int32  `json:"episode"`
}

type Invoice struct {
	ChatID    int64
	MessageID int
	Year      int32
	Payload   InvoicePayload
}

func (i Invoice) GenVideoOwnershipMsg(ctx context.Context) (*bot.SendMessageParams, error) {
	isUserHasVideo, err := retry.DoWithData(
		func() (bool, error) {
			return services.Queries.CheckVideoOwnership(ctx, repository.CheckVideoOwnershipParams{
				UserID:                 i.Payload.UserID,
				RunningManVideoEpisode: i.Payload.Episode,
			})
		},
		retry.Attempts(3),
		retry.LastErrorOnly(true),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to check video ownership: %w", err)
	}

	if isUserHasVideo {
		text := fmt.Sprintf("Ooops... kamu tidak bisa membeli video Running Man episode %d karena kamu telah memilikinya.", i.Payload.Episode)
		msg := bot.SendMessageParams{
			ChatID: i.ChatID,
			Text:   text,
		}

		return &msg, nil
	}

	return nil, nil
}

func (i Invoice) GenInvoiceExpirationMsg(ctx context.Context) (*bot.SendMessageParams, error) {
	isInvoiceUnexpired, err := retry.DoWithData(
		func() (bool, error) {
			return services.Queries.CheckInvoiceExpiration(ctx, repository.CheckInvoiceExpirationParams{
				UserID:                 i.Payload.UserID,
				RunningManVideoEpisode: i.Payload.Episode,
			})
		},
		retry.Attempts(3),
		retry.LastErrorOnly(true),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to check invoice expiration: %w", err)
	}

	if isInvoiceUnexpired {
		text := fmt.Sprintf(
			"Tidak dapat membuat tagihan untuk pembelian Running Man episode %d karena kamu telah memiliki tagihan yang masih valid.\n\nGunakan tagihan tersebut untuk melakukan pembayaran.",
			i.Payload.Episode,
		)

		msg := bot.SendMessageParams{
			ChatID: i.ChatID,
			Text:   text,
		}

		return &msg, nil
	}

	return nil, nil
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
		ChatID:      i.ChatID,
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
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Year:      int32(year),
		Payload: InvoicePayload{
			ID:      invoiceUUID.String(),
			UserID:  update.CallbackQuery.From.ID,
			Episode: int32(episode),
		},
	}

	videoOwnershipMsg, err := i.GenVideoOwnershipMsg(ctx)
	if err != nil {
		logger.Err(err).Send()
		return
	}

	invoiceExpirationMsg, err := i.GenInvoiceExpirationMsg(ctx)
	if err != nil {
		logger.Err(err).Send()
		return
	}

	if videoOwnershipMsg != nil || invoiceExpirationMsg != nil {
		var msg *bot.SendMessageParams
		if videoOwnershipMsg != nil {
			msg = videoOwnershipMsg
		} else {
			msg = invoiceExpirationMsg
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
