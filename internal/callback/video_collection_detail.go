package callback

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/avast/retry-go/v4"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mdayat/running-man/configs/env"
	"github.com/mdayat/running-man/configs/services"
	"github.com/mdayat/running-man/repository"
)

var (
	TypeVideoCollectionDetail InlineKeyboardType = "video_collection_detail"
	TypeVideoLink             InlineKeyboardType = "video_link"
)

type VideoCollectionDetail struct {
	ChatID          int64
	MessageID       int
	Episode         int32
	IsTypeVideoLink bool
}

func (vcd VideoCollectionDetail) GenInlineKeyboard(inlineKeyboardType InlineKeyboardType) tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(tg.NewInlineKeyboardRow(
		tg.NewInlineKeyboardButtonData("Buat Tautan", fmt.Sprintf("%s:%d", inlineKeyboardType, vcd.Episode)),
		tg.NewInlineKeyboardButtonData("Tidak", fmt.Sprintf("%s:%s", TypeVideoCollection, "")),
	))
}

func (vcd VideoCollectionDetail) GenVideoLink(libraryID int64, videoID string, year int32) string {
	TOKEN_AUTH_KEY := os.Getenv(fmt.Sprintf("TOKEN_AUTH_KEY_%d", year))
	duration := time.Now().Add(time.Minute * 3)
	durationUnix := duration.Unix()

	concatenatedString := TOKEN_AUTH_KEY + videoID + fmt.Sprintf("%d", durationUnix)
	hash := sha256.New()
	hash.Write([]byte(concatenatedString))
	videoLinkToken := hex.EncodeToString(hash.Sum(nil))

	return fmt.Sprintf("%s/%d/%s?token=%s&expires=%d", env.DIRECT_EMBED_BASE_URL, libraryID, videoID, videoLinkToken, durationUnix)
}

func (vcd VideoCollectionDetail) Process() (tg.Chattable, error) {
	var chat tg.Chattable
	if vcd.IsTypeVideoLink {
		retryFunc := func() (repository.GetRunningManVideoAndLibraryByEpisodeRow, error) {
			return services.Queries.GetRunningManVideoAndLibraryByEpisode(context.TODO(), vcd.Episode)
		}

		result, err := retry.DoWithData(retryFunc, retry.Attempts(3))
		if err != nil {
			return nil, fmt.Errorf("failed to get running man video and library by episode: %w", err)
		}

		videoID, err := result.RunningManVideoID.Value()
		if err != nil {
			return nil, fmt.Errorf("failed to get video UUID from pgtype.UUID: %w", err)
		}

		videoLink := vcd.GenVideoLink(result.RunningManLibraryID, fmt.Sprintf("%s", videoID), result.Year)
		text := "Tautan telah dibuat, klik tombol \"Tonton\" untuk menontonnya.\n\nTautan hanya berlaku selama tiga menit. Setelah itu, tautan menjadi invalid dan tidak dapat diakses.\n\nMeskipun tautan invalid, kamu tetap bisa menonton selama kamu tidak meninggalkan atau me-refresh browser."

		msg := tg.NewMessage(vcd.ChatID, text)
		msg.ReplyMarkup = tg.NewInlineKeyboardMarkup(tg.NewInlineKeyboardRow(tg.NewInlineKeyboardButtonURL("Tonton", videoLink)))
		chat = msg
	} else {
		text := fmt.Sprintf(`Tombol "Buat Tautan" akan membuat tautan untuk menonton video Running Man episode %d. Apakah kamu ingin membuat tautan?`, vcd.Episode)
		chat = tg.NewEditMessageTextAndMarkup(vcd.ChatID, vcd.MessageID, text, vcd.GenInlineKeyboard(TypeVideoLink))
	}

	return chat, nil
}
