package callback

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/mdayat/running-man/configs/env"
	"github.com/mdayat/running-man/configs/services"
	"github.com/mdayat/running-man/repository"
	"github.com/rs/zerolog/log"
)

var TypeVideoLink = "video_link"

type VideoLink struct {
	ChatID    int64
	MessageID int
	Episode   int32
}

func (vl VideoLink) GenVideoLinkMsg(ctx context.Context) (*bot.SendMessageParams, error) {
	result, err := retry.DoWithData(
		func() (repository.GetRunningManVideoAndLibraryByEpisodeRow, error) {
			return services.Queries.GetRunningManVideoAndLibraryByEpisode(ctx, vl.Episode)
		},
		retry.Attempts(3),
		retry.LastErrorOnly(true),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get running man video and library by episode: %w", err)
	}

	tokenAuthKey := os.Getenv(fmt.Sprintf("TOKEN_AUTH_KEY_%d", result.Year))
	duration := time.Now().Add(time.Minute * 3)
	durationUnix := duration.Unix()

	videoID, err := result.RunningManVideoID.Value()
	if err != nil {
		return nil, fmt.Errorf("failed to get video UUID from pgtype.UUID: %w", err)
	}

	concatenatedString := tokenAuthKey + fmt.Sprintf("%s", videoID) + fmt.Sprintf("%d", durationUnix)
	hash := sha256.New()
	hash.Write([]byte(concatenatedString))
	videoLinkToken := hex.EncodeToString(hash.Sum(nil))

	url := fmt.Sprintf("%s/%d/%s?token=%s&expires=%d", env.DirectEmbedBaseURL, result.RunningManLibraryID, videoID, videoLinkToken, durationUnix)
	text := fmt.Sprintf(
		"Tautan untuk video Running Man episode %d telah dibuat, klik tombol \"Tonton\" untuk menontonnya.\n\nTautan hanya berlaku selama tiga menit. Setelah itu, tautan menjadi invalid dan tidak dapat diakses.\n\nMeskipun tautan invalid, kamu tetap bisa menonton selama kamu tidak meninggalkan atau me-refresh browser.",
		vl.Episode,
	)

	msg := bot.SendMessageParams{
		ChatID: vl.ChatID,
		Text:   text,
		ReplyMarkup: &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{{{Text: "Tonton", URL: url}}},
		},
	}

	return &msg, nil
}

func VideoLinkHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	logger := log.Ctx(ctx).With().Logger()
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	episode, err := strconv.Atoi(strings.Split(update.CallbackQuery.Data, ":")[1])
	if err != nil {
		logger.Err(err).Msg("failed to convert episode string to int")
		return
	}

	vl := VideoLink{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Episode:   int32(episode),
	}

	videoLinkMsg, err := vl.GenVideoLinkMsg(ctx)
	if err != nil {
		logger.Err(err).Msg("failed to generate video link message")
		return
	}

	_, err = retry.DoWithData(
		func() (*models.Message, error) {
			return b.SendMessage(ctx, videoLinkMsg)
		},
		retry.Attempts(3),
		retry.LastErrorOnly(true),
	)

	if err != nil {
		logger.Err(err).Msgf("failed to send %s callback edit message", TypeVideoLink)
		return
	}
}
