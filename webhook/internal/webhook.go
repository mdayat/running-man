package internal

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"webhook/configs/env"
	"webhook/configs/services"
	"webhook/repository"

	"github.com/avast/retry-go/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
)

func createSignature(bytes []byte) string {
	key := []byte(env.TripayPrivateKey)
	hash := hmac.New(sha256.New, key)
	hash.Write(bytes)

	return hex.EncodeToString(hash.Sum(nil))
}

func sendTelegramMessage(userID int64, text string) error {
	payload := struct {
		ChatID string `json:"chat_id"`
		Text   string `json:"text"`
	}{
		ChatID: fmt.Sprintf("%d", userID),
		Text:   text,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal message payload")
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", env.BotToken)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to send POST request")
	}
	defer resp.Body.Close()

	return nil
}

type processSuccessfulPaymentParams struct {
	reference       string
	userID          int64
	merchantRefUUID uuid.UUID
	totalAmount     int32
	status          string
}

func processSuccessfulPayment(ctx context.Context, arg processSuccessfulPaymentParams) (err error) {
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
	err = qtx.CreatePayment(ctx, repository.CreatePaymentParams{
		ID:         arg.reference,
		UserID:     arg.userID,
		InvoiceID:  pgtype.UUID{Bytes: arg.merchantRefUUID, Valid: true},
		AmountPaid: int32(arg.totalAmount),
		Status:     arg.status,
	})
	if err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}

	oneMonth := (time.Hour * 24) * 30
	err = qtx.UpdateUserSubscription(ctx, repository.UpdateUserSubscriptionParams{
		ID:                    arg.userID,
		SubscriptionExpiredAt: pgtype.Timestamptz{Time: time.Now().Add(oneMonth), Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to update user subscription: %w", err)
	}

	text := "Pembayaran berhasil. Kamu dapat menonton video Running Man sepuasnya selama satu bulan."
	if err = sendTelegramMessage(arg.userID, text); err != nil {
		return fmt.Errorf("failed to send telegram message: %w", err)
	}

	return nil
}

type processFailedPaymentParams struct {
	reference       string
	userID          int64
	merchantRefUUID uuid.UUID
	totalAmount     int32
	status          string
}

func processFailedPayment(ctx context.Context, arg processFailedPaymentParams) (err error) {
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
	err = qtx.CreatePayment(ctx, repository.CreatePaymentParams{
		ID:         arg.reference,
		UserID:     arg.userID,
		InvoiceID:  pgtype.UUID{Bytes: arg.merchantRefUUID, Valid: true},
		AmountPaid: int32(arg.totalAmount),
		Status:     arg.status,
	})
	if err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}

	text := "Pembayaran gagal. Silakan hubungi tim dukungan kami melalui perintah /support."
	if err = sendTelegramMessage(arg.userID, text); err != nil {
		return fmt.Errorf("failed to send telegram message: %w", err)
	}

	return nil
}

type requestBody struct {
	Reference   string `json:"reference"`
	MerchantRef string `json:"merchant_ref"`
	TotalAmount int    `json:"total_amount"`
	Status      string `json:"status"`
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log.Ctx(ctx).With().Logger()

	logger.Info().Msg("reading tripay webhook request...")
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error().Err(err).Caller().Int("status_code", http.StatusInternalServerError).Msg("failed to read tripay webhook request")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	logger.Info().Msg("successfully read tripay webhook request")

	signature := createSignature(bytes)
	tripaySignature := r.Header.Get("X-Callback-Signature")

	logger.Info().Msg("validating tripay signature...")
	if signature != tripaySignature {
		logger.Error().Err(err).Caller().Int("status_code", http.StatusForbidden).Msg("failed to validate tripay signature")
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}
	logger.Info().Msg("successfully validated tripay signature")

	var body requestBody
	logger.Info().Msg("unmarshalling tripay webhook request body...")
	if err := json.Unmarshal(bytes, &body); err != nil {
		logger.Error().Err(err).Caller().Int("status_code", http.StatusInternalServerError).Msg("failed to unmarshal tripay webhook request body")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	logger.Info().Msg("successfully unmarshalled tripay webhook request body")

	logger.Info().Msg("parsing merchant ref string to UUID...")
	merchantRefUUID, err := uuid.Parse(body.MerchantRef)
	if err != nil {
		logger.Error().Err(err).Caller().Int("status_code", http.StatusInternalServerError).Msg("failed to parse merchant ref string to UUID")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	logger.Info().Msg("successfully parsed merchant ref string to UUID")

	logger.Info().Msg("fetching user ID by invoice ID...")
	userID, err := retry.DoWithData(
		func() (int64, error) {
			return services.Queries.GetUserIDByInvoiceID(ctx, pgtype.UUID{Bytes: merchantRefUUID, Valid: true})
		},
		retry.Attempts(3),
		retry.LastErrorOnly(true),
	)
	if err != nil {
		logger.Error().Err(err).Caller().Int("status_code", http.StatusInternalServerError).Msg("failed to fetch user ID by invoice ID")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	logger.Info().Msg("successfully fetched user ID by invoice ID")

	if body.Status == "PAID" {
		logger.Info().Msg("processing successful payment...")
		err = processSuccessfulPayment(ctx, processSuccessfulPaymentParams{
			reference:       body.Reference,
			userID:          userID,
			merchantRefUUID: merchantRefUUID,
			totalAmount:     int32(body.TotalAmount),
			status:          body.Status,
		})
		if err != nil {
			logger.Error().Err(err).Caller().Int("status_code", http.StatusInternalServerError).Msg("failed to process successful payment")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		logger.Info().Msg("successfully processed successful payment")
	} else {
		logger.Info().Msg("processing failed payment...")
		err = processFailedPayment(ctx, processFailedPaymentParams{
			reference:       body.Reference,
			userID:          userID,
			merchantRefUUID: merchantRefUUID,
			totalAmount:     int32(body.TotalAmount),
			status:          body.Status,
		})
		if err != nil {
			logger.Error().Err(err).Caller().Int("status_code", http.StatusInternalServerError).Msg("failed to process failed payment")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		logger.Info().Msg("successfully processed failed payment")
	}

	payload := struct {
		Status bool `json:"status"`
	}{Status: true}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	logger.Info().Msg("sending successful tripay webhook request...")
	if err = json.NewEncoder(w).Encode(payload); err != nil {
		logger.Error().Err(err).Caller().Int("status_code", http.StatusInternalServerError).Msg("failed to send successful tripay webhook request")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	logger.Info().Msg("successfully sent tripay webhook request")
}
