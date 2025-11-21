package utils

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	// "io"
	"net/http"
	"net/url"
	// "regexp"
	"strings"
	"time"
)

// Struktur Request ke Internal API YouTube
type InnerTubeRequest struct {
	Context InnerTubeContext `json:"context"`
	VideoID string           `json:"videoId"`
}

type InnerTubeContext struct {
	Client InnerTubeClient `json:"client"`
}

type InnerTubeClient struct {
	Hl            string `json:"hl"`
	Gl            string `json:"gl"`
	ClientName    string `json:"clientName"`
	ClientVersion string `json:"clientVersion"`
}

// Struktur Response (Kita ambil bagian caption aja)
type InnerTubeResponse struct {
	Captions struct {
		PlayerCaptionsTracklistRenderer struct {
			CaptionTracks []struct {
				BaseUrl      string `json:"baseUrl"`
				Name         struct { SimpleText string `json:"simpleText"` } `json:"name"`
				LanguageCode string `json:"languageCode"`
				Kind         string `json:"kind"`
			} `json:"captionTracks"`
		} `json:"playerCaptionsTracklistRenderer"`
	} `json:"captions"`
	VideoDetails struct {
		Title string `json:"title"`
	} `json:"videoDetails"`
	PlayabilityStatus struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	} `json:"playabilityStatus"`
}

// Struktur XML Transkrip (Output Akhir)
type TranscriptXML struct {
	Text []struct {
		Body string `xml:",chardata"`
	} `xml:"text"`
}

func ExtractTextFromYouTube(videoURL string) (string, error) {
	videoID, err := extractVideoID(videoURL)
	if err != nil {
		return "", err
	}

	// 1. Setup Payload JSON (Pura-pura jadi Web Client)
	// Ini endpoint resmi yang dipake website YouTube
	payload := InnerTubeRequest{
		VideoID: videoID,
		Context: InnerTubeContext{
			Client: InnerTubeClient{
				ClientName:    "WEB",
				ClientVersion: "2.20230920.00.00", // Versi client yang stabil
				Hl:            "en",
				Gl:            "US",
			},
		},
	}

	jsonData, _ := json.Marshal(payload)

	// 2. Tembak API YouTube (Innertube)
	// Key ini adalah Public Key YouTube Web (selalu sama dan publik)
	apiUrl := "https://www.youtube.com/youtubei/v1/player?key=AIzaSyDHEQWpBthrtuBhgUnVW3MkIvwfTPmBnQ8"
	
	req, _ := http.NewRequest("POST", apiUrl, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("gagal koneksi ke youtube api: %w", err)
	}
	defer resp.Body.Close()

	// 3. Parsing JSON Response
	var data InnerTubeResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("gagal decode json response: %w", err)
	}

	// Cek Status Video
	if data.PlayabilityStatus.Status != "OK" {
		return "", fmt.Errorf("video tidak bisa diputar: %s", data.PlayabilityStatus.Reason)
	}

	tracks := data.Captions.PlayerCaptionsTracklistRenderer.CaptionTracks
	if len(tracks) == 0 {
		return "", errors.New("tidak ada caption/subtitle tersedia untuk video ini")
	}

	// 4. Pilih Bahasa (Indo -> Inggris -> Auto)
	var selectedURL string

	// Cari Indo
	for _, t := range tracks {
		if strings.HasPrefix(t.LanguageCode, "id") {
			selectedURL = t.BaseUrl
			break
		}
	}
	// Cari Inggris
	if selectedURL == "" {
		for _, t := range tracks {
			if strings.HasPrefix(t.LanguageCode, "en") {
				selectedURL = t.BaseUrl
				break
			}
		}
	}
	// Fallback
	if selectedURL == "" {
		selectedURL = tracks[0].BaseUrl
	}

	// 5. Download XML Transkrip
	// Kita pakai URL yang didapat dari API, biasanya aman didownload langsung
	reqDl, _ := http.NewRequest("GET", selectedURL, nil)
	reqDl.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	
	respDl, err := client.Do(reqDl)
	if err != nil {
		return "", fmt.Errorf("gagal download file subtitle: %w", err)
	}
	defer respDl.Body.Close()

	if respDl.StatusCode != 200 {
		return "", fmt.Errorf("gagal download subtitle, status: %d", respDl.StatusCode)
	}

	// 6. Parse XML ke Text
	var transcriptXML TranscriptXML
	if err := xml.NewDecoder(respDl.Body).Decode(&transcriptXML); err != nil {
		return "", fmt.Errorf("gagal parsing xml subtitle: %w", err)
	}

	var fullText strings.Builder
	for _, item := range transcriptXML.Text {
		// Decode HTML entities (misal &#39; jadi ')
		text := html.UnescapeString(item.Body)
		text = strings.TrimSpace(text)
		if text != "" {
			fullText.WriteString(text)
			fullText.WriteString(" ")
		}
	}

	final := strings.TrimSpace(fullText.String())
	if len(final) < 10 {
		return "", errors.New("hasil transkrip kosong")
	}

	return final, nil
}

// --- FUNGSI GET TITLE (Pake API yang sama biar konsisten) ---
func GetVideoTitle(videoURL string) (string, error) {
	videoID, err := extractVideoID(videoURL)
	if err != nil { return "Unknown Title", nil }

	payload := InnerTubeRequest{
		VideoID: videoID,
		Context: InnerTubeContext{
			Client: InnerTubeClient{
				ClientName:    "WEB",
				ClientVersion: "2.20230920.00.00",
				Hl:            "en",
				Gl:            "US",
			},
		},
	}
	jsonData, _ := json.Marshal(payload)

	apiUrl := "https://www.youtube.com/youtubei/v1/player?key=AIzaSyDHEQWpBthrtuBhgUnVW3MkIvwfTPmBnQ8"
	req, _ := http.NewRequest("POST", apiUrl, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil { return "", err }
	defer resp.Body.Close()

	var data InnerTubeResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err == nil && data.VideoDetails.Title != "" {
		return data.VideoDetails.Title, nil
	}
	return "Unknown Title", nil
}

// Helper ID
func extractVideoID(videoURL string) (string, error) {
	u, err := url.Parse(videoURL)
	if err != nil { return "", err }
	if u.Host == "youtu.be" { return strings.TrimPrefix(u.Path, "/"), nil }
	q := u.Query()
	v := q.Get("v")
	if v == "" { return "", errors.New("no video id") }
	return v, nil
}