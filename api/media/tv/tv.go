package tv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/telebot.v3"
	"io"
	"log"
	"movie-manager-bot/api"
	"movie-manager-bot/api/media/image"
	"movie-manager-bot/config"
	"movie-manager-bot/helpers/utils"
	"net/http"
)

func GetTV(tvId int) (*TV, error) {
	url := utils.MakeUrl(fmt.Sprintf("%s/%v", config.Cfg.Endpoints.GetTv, tvId), nil)

	resp, err := api.Client.HttpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching tv data: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	var result TV
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error parsing json response: %w", err)
	}

	return &result, nil
}

func ShowTV(context telebot.Context, tvData *TV) error {
	imgBuffer, err := image.GetImage(tvData.PosterPath)
	if err != nil {
		log.Printf("Error retrieving image: %v", err)
		return fmt.Errorf("could not retrieve image: %w", err)
	}

	caption := fmt.Sprintf(
		"📺 *Name*: %v\n"+
			"📝 *Overview*: %v\n"+
			"📜 *Status*: %v\n"+
			"🔞 *Is Adult*: %v\n"+
			"🔥 *Popularity*: %v\n"+
			"🎥 *Seasons*: %v\n"+
			"#️⃣ *Episodes*: %v\n\n",
		tvData.Name,
		tvData.Overview,
		tvData.Status,
		tvData.Adult,
		tvData.Popularity,
		tvData.Seasons,
		tvData.Episodes,
	)

	backBtn := &telebot.ReplyMarkup{}
	backBtn.Inline(
		backBtn.Row(backBtn.Data("🔙 Back to list", "tv|back_to_pagination|")),
	)

	err = context.Delete()
	if err != nil {
		log.Printf("Failed to delete original message: %v", err)
	}

	imageFile := &telebot.Photo{
		File:    telebot.File{FileReader: bytes.NewReader(imgBuffer.Bytes())},
		Caption: caption,
	}

	_, err = context.Bot().Send(context.Chat(), imageFile, backBtn, telebot.ModeMarkdown)
	if err != nil {
		log.Printf("Failed to send tv details: %v", err)
		return fmt.Errorf("could not send tv details: %w", err)
	}

	log.Printf("Tv details successfully sent for tv ID: %d", tvData.Id)
	return nil
}
