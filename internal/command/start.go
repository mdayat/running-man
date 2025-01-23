package command

import (
	"context"

	"github.com/avast/retry-go/v4"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/mdayat/running-man/configs/services"
	"github.com/mdayat/running-man/repository"
	"github.com/rs/zerolog/log"
)

func StartHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	logger := log.Ctx(ctx).With().Logger()
	isUserExist, err := retry.DoWithData(
		func() (bool, error) {
			return services.Queries.CheckUserExistence(ctx, update.Message.From.ID)
		},
		retry.Attempts(3),
		retry.LastErrorOnly(true),
	)

	if err != nil {
		logger.Err(err).Msg("failed to check user existence")
		return
	}

	if !isUserExist {
		err := retry.Do(
			func() error {
				return services.Queries.CreateUser(ctx, repository.CreateUserParams{
					ID:        update.Message.From.ID,
					FirstName: update.Message.From.FirstName,
				})
			},
			retry.Attempts(3),
			retry.LastErrorOnly(true),
		)

		if err != nil {
			logger.Err(err).Msg("failed to create user")
			return
		}
	}

	text := `
	Selamat datang di @RunningManSeriesBot! Dengan bot ini, kamu dapat mengelola dan membeli episode lama dari "Running Man." Berikut adalah perintah yang dapat kamu gunakan:

	/start - Pengenalan dan daftar perintah yang tersedia
	/browse - Jelajahi episode Running Man
	/collection - Lihat koleksi video Running Man yang kamu miliki
	/support - Butuh bantuan? Dapatkan dukungan teknis, tips, dan jawaban atas pertanyaan seputar penggunaan aplikasi kami
	`

	_, err = retry.DoWithData(
		func() (*models.Message, error) {
			return b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   text,
			})
		},
		retry.Attempts(3),
		retry.LastErrorOnly(true),
	)

	if err != nil {
		logger.Err(err).Msg("failed to send start command message")
		return
	}
}
