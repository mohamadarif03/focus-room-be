package utils

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var youtubeService *youtube.Service

func InitYouTubeService(apiKey string) error {
	if apiKey == "" {
		return errors.New("YOUTUBE_API_KEY kosong")
	}

	ctx := context.Background()
	var err error
	youtubeService, err = youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return fmt.Errorf("gagal membuat YouTube service: %w", err)
	}
	return nil
}

func ExtractTextFromYouTube(youtubeURL string) (string, error) {
	if youtubeService == nil {
		return "", errors.New("YouTube service belum diinisialisasi")
	}

	videoID, err := getVideoID(youtubeURL)
	if err != nil {
		return "", err
	}

	ctx := context.Background()
	call := youtubeService.Captions.List([]string{"snippet"}, videoID).Context(ctx)
	response, err := call.Do()
	if err != nil {
		return "", fmt.Errorf("gagal mengambil daftar caption: %w", err)
	}

	if len(response.Items) == 0 {
		return "", errors.New("video ini tidak memiliki caption/transkrip")
	}

	var captionTrack *youtube.Caption
	for _, item := range response.Items {
		lang := item.Snippet.Language
		if lang == "id" {
			captionTrack = item
			break
		}
		if lang == "en" {
			captionTrack = item
		}
	}

	if captionTrack == nil {
		return "", errors.New("tidak ditemukan transkrip 'id' atau 'en'")
	}

	downloadURL := fmt.Sprintf("https://www.youtube.com/api/timedtext?v=%s&lang=%s&fmt=srv3", videoID, captionTrack.Snippet.Language)
	resp, err := http.Get(downloadURL)
	if err != nil {
		return "", fmt.Errorf("gagal download transkrip: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("gagal baca body transkrip: %w", err)
	}

	xmlData := string(body)
	var allText strings.Builder
	lines := strings.Split(xmlData, "</text>")
	for _, line := range lines {
		start := strings.Index(line, `">`)
		if start != -1 {
			text := line[start+2:]
			text = strings.ReplaceAll(text, "&#39;", "'")
			text = strings.ReplaceAll(text, "&amp;", "&")
			allText.WriteString(text)
			allText.WriteString(" ")
		}
	}

	return allText.String(), nil
}

func getVideoID(youtubeURL string) (string, error) {
	u, err := url.Parse(youtubeURL)
	if err != nil {
		return "", err
	}
	if u.Host == "www.youtube.com" || u.Host == "youtube.com" {
		videoID := u.Query().Get("v")
		if videoID != "" {
			return videoID, nil
		}
	}
	if u.Host == "youtu.be" {
		if len(u.Path) > 1 {
			return u.Path[1:], nil
		}
	}
	return "", errors.New("tidak bisa parsing video ID dari URL YouTube")
}
