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
	TypeEpisode  InlineKeyboardType = "episode"
	TypePurchase InlineKeyboardType = "purchase"
)

type RunningManEpisode struct {
	UserID       int64
	ChatID       int64
	MessageID    int
	Year         int32
	Episode      int32
	IsPurchasing bool
}

func (rme RunningManEpisode) GenInlineKeyboard(inlineKeyboardType InlineKeyboardType) tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(tg.NewInlineKeyboardRow(
		tg.NewInlineKeyboardButtonData("Iya", fmt.Sprintf("%s:%d:%d", inlineKeyboardType, rme.Year, rme.Episode)),
		tg.NewInlineKeyboardButtonData("Tidak", fmt.Sprintf("%s:%d", TypeVideos, rme.Year)),
	))
}

type InvoicePayload struct {
	UserID  int64 `json:"user_id"`
	Episode int32 `json:"episode"`
	Amount  int
}

func (rme RunningManEpisode) GenInvoice() (tg.Chattable, error) {
	price, err := retry.DoWithData(
		func() (_ int32, err error) {
			return services.Queries.GetRunningManVideoPrice(context.TODO(), rme.Episode)
		},
		retry.Attempts(3),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get running man video price: %w", err)
	}

	tax := int(math.Ceil(float64(price) * 0.11))
	priceAfterTax := int(price) + tax

	payload := InvoicePayload{
		UserID:  rme.UserID,
		Episode: rme.Episode,
		Amount:  priceAfterTax,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal invoice payload: %w", err)
	}

	invoice := tg.NewInvoice(
		rme.ChatID,
		fmt.Sprintf("Running Man Episode %d", rme.Episode),
		fmt.Sprintf("Pembelian Running Man episode %d (%d), harga termasuk pajak.", rme.Episode, rme.Year),
		string(payloadJSON),
		"",
		"start_param_unique_v1",
		"XTR",
		[]tg.LabeledPrice{{Label: "XTR", Amount: payload.Amount}},
	)
	invoice.SuggestedTipAmounts = []int{}

	return invoice, nil
}

func (rme RunningManEpisode) Process() (_ tg.Chattable, err error) {
	var chat tg.Chattable
	if rme.IsPurchasing == true {
		isUserHasVideo, err := retry.DoWithData(
			func() (_ bool, err error) {
				return services.Queries.CheckUserVideo(context.TODO(), repository.CheckUserVideoParams{
					UserID:                 rme.UserID,
					RunningManVideoEpisode: rme.Episode,
				})
			},
			retry.Attempts(3),
		)

		if isUserHasVideo {
			text := fmt.Sprintf("Ooops... kamu tidak bisa membeli video Running Man episode %d karena kamu telah memilikinya.", rme.Episode)
			chat := tg.NewMessage(rme.ChatID, text)
			return chat, nil
		}

		chat, err = rme.GenInvoice()
		if err != nil {
			return nil, fmt.Errorf("failed to generate invoice for running man episode: %w", err)
		}
	} else {
		text := fmt.Sprintf("Apakah kamu ingin membeli Running Man episode %d?", rme.Episode)
		chat = tg.NewEditMessageTextAndMarkup(rme.ChatID, rme.MessageID, text, rme.GenInlineKeyboard(TypePurchase))
	}

	return chat, nil
}
