package command

import (
	"context"
	"fmt"

	"github.com/avast/retry-go/v4"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mdayat/running-man/configs/services"
	"github.com/mdayat/running-man/repository"
)

var defaultMessage = `
Selamat datang di @RunningManSeriesBot! Dengan bot ini, Anda dapat mengelola dan membeli episode lama dari "Running Man." Berikut adalah perintah yang dapat Anda gunakan:

/start - Selamat datang dan pengenalan
/help - Dapatkan bantuan dan daftar perintah yang tersedia
/browse - Tampilkan katalog episode
/search [nomor_episode] - Cari episode tertentu
/buy [nomor_episode] - Beli episode
/checkout - Selesaikan pembelian dengan Telegram stars
/mycollection - Tampilkan episode yang sudah dibeli
/feedback - Kirimkan umpan balik atau laporkan masalah
`

type Default struct {
	UserID    int64
	ChatID    int64
	FirstName string
}

func (d Default) Process() (tg.Chattable, error) {
	isUserExist, err := retry.DoWithData(
		func() (_ bool, err error) {
			return services.Queries.CheckUserExistence(context.TODO(), d.UserID)
		},
		retry.Attempts(3),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}

	if !isUserExist {
		err := retry.Do(
			func() error {
				return services.Queries.CreateUser(context.TODO(), repository.CreateUserParams{
					ID:        d.UserID,
					FirstName: d.FirstName,
				})
			},
			retry.Attempts(3),
		)

		if err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	}

	chat := tg.NewMessage(d.ChatID, defaultMessage)
	return chat, nil
}
