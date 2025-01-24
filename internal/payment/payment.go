package payment

import (
	"context"
	"fmt"

	"github.com/avast/retry-go/v4"
	badger "github.com/dgraph-io/badger/v4"
	"github.com/go-telegram/bot"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mdayat/running-man/configs/services"
	"github.com/mdayat/running-man/repository"
)

type CheckVideoOwnershipParams struct {
	UserID  int64
	Episode int32
}

func CheckVideoOwnership(ctx context.Context, arg CheckVideoOwnershipParams) (bool, error) {
	return retry.DoWithData(
		func() (bool, error) {
			return services.Queries.CheckVideoOwnership(ctx, repository.CheckVideoOwnershipParams{
				UserID:                 arg.UserID,
				RunningManVideoEpisode: arg.Episode,
			})
		},
		retry.Attempts(3),
		retry.LastErrorOnly(true),
	)
}

type CheckInvoiceExpirationParams struct {
	UserID  int64
	Episode int32
}

func CheckInvoiceExpiration(ctx context.Context, arg CheckInvoiceExpirationParams) (bool, error) {
	return retry.DoWithData(
		func() (bool, error) {
			return services.Queries.CheckInvoiceExpiration(ctx, repository.CheckInvoiceExpirationParams{
				UserID:                 arg.UserID,
				RunningManVideoEpisode: arg.Episode,
			})
		},
		retry.Attempts(3),
		retry.LastErrorOnly(true),
	)
}

func CheckExpiredInvoice(ctx context.Context, invoiceID string) (bool, error) {
	invoiceUUID, err := uuid.Parse(invoiceID)
	if err != nil {
		return false, fmt.Errorf("failed to convert invoice ID string to UUID: %w", err)
	}

	return retry.DoWithData(
		func() (bool, error) {
			return services.Queries.CheckExpiredInvoice(ctx, pgtype.UUID{Bytes: invoiceUUID, Valid: true})
		},
		retry.Attempts(3),
		retry.LastErrorOnly(true),
	)
}

func MakeVideoOwnershipText(episode int32) string {
	return fmt.Sprintf("Ooops... kamu tidak bisa membeli video Running Man episode %d karena kamu telah memilikinya.", episode)
}

func MakeUnexpiredInvoiceText(episode int32) string {
	return fmt.Sprintf(
		"Tidak dapat membuat tagihan untuk pembelian Running Man episode %d karena kamu memiliki tagihan yang masih valid.\n\nGunakan tagihan tersebut untuk melakukan pembayaran.",
		episode,
	)
}

func MakeExpiredInvoiceText(episode int32) string {
	return fmt.Sprintf(
		"Kamu tidak memiliki tagihan yang valid untuk melakukan pembelian video Running Man episode %d. Silakan buat tagihan terlebih dahulu.",
		episode,
	)
}

type InvoicePayload struct {
	ID      string `json:"id"`
	ChatID  int64  `json:"chat_id"`
	UserID  int64  `json:"user_id"`
	Episode int32  `json:"episode"`
}

type InsertAndSendSuccessfulPaymentParams struct {
	Payload   InvoicePayload
	PaymentID string
}

func InsertAndSendSuccessfulPayment(ctx context.Context, b *bot.Bot, arg InsertAndSendSuccessfulPaymentParams) (err error) {
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

		videoCollectionKey := fmt.Sprintf("%d:%s", arg.Payload.UserID, "video_collection")
		if err := services.Badger.Update(func(txn *badger.Txn) error { return txn.Delete([]byte(videoCollectionKey)) }); err != nil {
			return fmt.Errorf("failed to invalidate %s key: %w", videoCollectionKey, err)
		}

		invoiceUUID, err := uuid.Parse(arg.Payload.ID)
		if err != nil {
			return fmt.Errorf("failed to convert invoice ID string to UUID: %w", err)
		}

		qtx := services.Queries.WithTx(tx)
		err = qtx.CreatePayment(ctx, repository.CreatePaymentParams{
			ID:        arg.PaymentID,
			UserID:    arg.Payload.UserID,
			InvoiceID: pgtype.UUID{Bytes: invoiceUUID, Valid: true},
		})

		if err != nil {
			return fmt.Errorf("failed to insert successful payment: %w", err)
		}

		err = qtx.CreateVideoCollection(ctx, repository.CreateVideoCollectionParams{
			UserID:                 arg.Payload.UserID,
			RunningManVideoEpisode: arg.Payload.Episode,
		})

		if err != nil {
			return fmt.Errorf("failed to insert video collection: %w", err)
		}

		text := fmt.Sprintf("Pembayaran untuk video Running Man episode %d telah berhasil. Tonton video baru kamu melalui perintah /collection.", arg.Payload.Episode)
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: arg.Payload.ChatID,
			Text:   text,
		})

		if err != nil {
			return fmt.Errorf("failed to send successful payment: %w", err)
		}

		return nil
	}

	return retry.Do(retryFunc, retry.Attempts(3), retry.LastErrorOnly(true))
}
