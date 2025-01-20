package callback

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"github.com/avast/retry-go/v4"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

type InvoicePayload struct {
	UserID  int64 `json:"user_id"`
	Episode int32 `json:"episode"`
	Amount  int
}

func (rmv RunningManVideo) GenInvoice() (tg.Chattable, error) {
	price, err := retry.DoWithData(
		func() (_ int32, err error) {
			return services.Queries.GetRunningManVideoPrice(context.TODO(), rmv.Episode)
		},
		retry.Attempts(3),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get running man video price: %w", err)
	}

	tax := int(math.Ceil(float64(price) * 0.11))
	priceAfterTax := int(price) + tax

	payload := InvoicePayload{
		UserID:  rmv.UserID,
		Episode: rmv.Episode,
		Amount:  priceAfterTax,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal invoice payload: %w", err)
	}

	invoice := tg.NewInvoice(
		rmv.ChatID,
		fmt.Sprintf("Running Man Episode %d", rmv.Episode),
		fmt.Sprintf("Pembelian Running Man episode %d (%d), harga termasuk pajak.", rmv.Episode, rmv.Year),
		string(payloadJSON),
		"",
		"start_param_unique_v1",
		"XTR",
		[]tg.LabeledPrice{{Label: "XTR", Amount: payload.Amount}},
	)
	invoice.SuggestedTipAmounts = []int{}

	return invoice, nil
}

func (rmv RunningManVideo) Process() (tg.Chattable, error) {
	var chat tg.Chattable
	if rmv.IsTypeInvoice == true {
		retryFunc := func() (_ bool, err error) {
			return services.Queries.CheckUserVideo(context.TODO(), repository.CheckUserVideoParams{
				UserID:                 rmv.UserID,
				RunningManVideoEpisode: rmv.Episode,
			})
		}

		isUserHasVideo, err := retry.DoWithData(retryFunc, retry.Attempts(3))
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
