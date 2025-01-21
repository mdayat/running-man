package callback

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/avast/retry-go/v4"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mdayat/running-man/configs/services"
	"github.com/mdayat/running-man/repository"
)

var (
	TypeVideo   InlineKeyboardType = "video"
	TypeInvoice InlineKeyboardType = "invoice"
)

type RunningManVideo struct {
	UserID        int64
	ChatID        int64
	MessageID     int
	Year          int32
	Episode       int32
	IsTypeInvoice bool
}

func (rmv RunningManVideo) GenInlineKeyboard(inlineKeyboardType InlineKeyboardType) tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(tg.NewInlineKeyboardRow(
		tg.NewInlineKeyboardButtonData("Iya", fmt.Sprintf("%s:%d:%d", inlineKeyboardType, rmv.Year, rmv.Episode)),
		tg.NewInlineKeyboardButtonData("Tidak", fmt.Sprintf("%s:%d", TypeVideos, rmv.Year)),
	))
}

type invoicePayload struct {
	ID      string `json:"id"`
	UserID  int64  `json:"user_id"`
	Episode int32  `json:"episode"`
}

func (rmv RunningManVideo) GenInvoice() (tg.Chattable, error) {
	isInvoiceUnexpired, err := retry.DoWithData(
		func() (bool, error) {
			return services.Queries.CheckInvoiceExpiration(context.TODO(), repository.CheckInvoiceExpirationParams{
				UserID:                 rmv.UserID,
				RunningManVideoEpisode: rmv.Episode,
			})
		},
		retry.Attempts(3),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to check invoice expiration: %w", err)
	}

	if isInvoiceUnexpired {
		text := fmt.Sprintf(
			"Tidak dapat membuat tagihan untuk pembelian Running Man episode %d karena kamu telah memiliki tagihan yang masih valid.\n\nGunakan tagihan tersebut untuk melakukan pembayaran.",
			rmv.Episode,
		)

		msg := tg.NewMessage(rmv.ChatID, text)
		return msg, nil
	}

	price, err := retry.DoWithData(
		func() (int32, error) {
			return services.Queries.GetRunningManVideoPrice(context.TODO(), rmv.Episode)
		},
		retry.Attempts(3),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get running man video price: %w", err)
	}

	tax := int(math.Ceil(float64(price) * 0.11))
	priceAfterTax := int(price) + tax

	invoiceUUID := uuid.New()
	payload := invoicePayload{
		ID:      invoiceUUID.String(),
		UserID:  rmv.UserID,
		Episode: rmv.Episode,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal invoice payload: %w", err)
	}

	title := fmt.Sprintf("Running Man Episode %d", rmv.Episode)
	description := fmt.Sprintf(
		"Pembelian Running Man episode %d (%d), harga sudah termasuk pajak. Segera lakukan pembayaran karena tagihan akan invalid setelah 1 jam.",
		rmv.Episode,
		rmv.Year,
	)

	prices := []tg.LabeledPrice{
		{
			Label:  title,
			Amount: priceAfterTax,
		},
	}

	invoice := tg.NewInvoice(
		rmv.ChatID,
		title,
		description,
		string(payloadJSON),
		"",
		"no_pay",
		"XTR",
		prices,
	)
	invoice.SuggestedTipAmounts = []int{}

	err = retry.Do(
		func() error {
			return services.Queries.CreateInvoice(context.TODO(), repository.CreateInvoiceParams{
				ID:                     pgtype.UUID{Bytes: invoiceUUID, Valid: true},
				UserID:                 payload.UserID,
				RunningManVideoEpisode: payload.Episode,
				Amount:                 int32(priceAfterTax),
				ExpiredAt:              pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
			})
		},
		retry.Attempts(3),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	return invoice, nil
}

func (rmv RunningManVideo) Process() (tg.Chattable, error) {
	var chat tg.Chattable
	if rmv.IsTypeInvoice {
		retryFunc := func() (bool, error) {
			return services.Queries.CheckUserVideo(context.TODO(), repository.CheckUserVideoParams{
				UserID:                 rmv.UserID,
				RunningManVideoEpisode: rmv.Episode,
			})
		}

		isUserHasVideo, err := retry.DoWithData(retryFunc, retry.Attempts(3))
		if err != nil {
			return nil, fmt.Errorf("failed to check user video: %w", err)
		}

		if isUserHasVideo {
			text := fmt.Sprintf("Ooops... kamu tidak bisa membeli video Running Man episode %d karena kamu telah memilikinya.", rmv.Episode)
			chat := tg.NewMessage(rmv.ChatID, text)
			return chat, nil
		}

		chat, err = rmv.GenInvoice()
		if err != nil {
			return nil, fmt.Errorf("failed to generate invoice for running man episode: %w", err)
		}
	} else {
		text := fmt.Sprintf("Apakah kamu ingin membeli Running Man episode %d?", rmv.Episode)
		chat = tg.NewEditMessageTextAndMarkup(rmv.ChatID, rmv.MessageID, text, rmv.GenInlineKeyboard(TypeInvoice))
	}

	return chat, nil
}
