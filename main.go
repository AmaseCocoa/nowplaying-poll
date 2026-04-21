package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/joho/godotenv"
	"github.com/mattn/go-mastodon"
	"go.etcd.io/bbolt"
)

type PlayingNowTrackMetadata struct {
	ArtistName  string `json:"artist_name"`
	ReleaseName string `json:"release_name"`
	TrackName   string `json:"track_name"`
}

type PlayingNowPayload struct {
	TrackMetadata PlayingNowTrackMetadata `json:"track_metadata"`
	PlayingNow    bool                    `json:"playing_now"`
}

type PlayingNowResponse struct {
	Payload struct {
		Listens []PlayingNowPayload `json:"listens"`
	} `json:"payload"`
}

type SocialSender struct {
	tmpl *template.Template
	mastodon *mastodon.Client
}

func sendPosts(track PlayingNowTrackMetadata, sender SocialSender) {
	data := map[string]string{
		"Track":  track.TrackName,
		"Artist": track.ArtistName,
		"Album": track.ReleaseName,
	}
	var doc bytes.Buffer
	sender.tmpl.Execute(&doc, data)
	text := doc.String()
	
	toot := &mastodon.Toot{
		Status: text,
	}
	_, err := sender.mastodon.PostStatus(context.Background(), toot)
	if err != nil {
		log.Fatal(err)
	}
}

func setupSender(db *bbolt.DB, template *template.Template) SocialSender {
	mastodonInstance := os.Getenv("MASTODON_INSTANCE")
	mstdn := MastodonSender {
		host: mastodonInstance,
		db: db,
	}
	
	
	return SocialSender {
		tmpl: template,
		mastodon: mstdn.GetToken(),
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
	    log.Fatal("Error loading .env file")
	}
		
	username := os.Getenv("LISTENBRAINZ_USER")
	url := fmt.Sprintf("https://api.listenbrainz.org/1/user/%s/playing-now", username)
	var tmpl *template.Template
	if postTemplate, exists := os.LookupEnv("SOCIAL_SENDER_FORMAT"); exists {
		tmpl, _ = template.New("socialSenderTmpl").Parse(postTemplate)
	} else {
		tmpl, _ = template.New("socialSenderTmpl").Parse("{{.Track}} - {{.Artist}} ({{.Album}})\n#NowPlaying")
	}
	
	db, _ := bbolt.Open("my.db", 0600, nil)
	defer db.Close()

	sender := setupSender(db, tmpl)
	
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	client := &http.Client{Timeout: 10 * time.Second}

	for range ticker.C {
		req, _ := http.NewRequest("GET", url, nil)

		req.Header.Set("User-Agent", "ListenBrainzWatcher/0.1.0 ( https://github.com/AmaseCocoa/nowplaying-poll )")
		if listenBrainzToken, exists := os.LookupEnv("LISTENBRAINZ_TOKEN"); exists {
			req.Header.Set("Authorization", fmt.Sprintf("Token %s", listenBrainzToken))
		}
		
		resp, err := client.Do(req)
		if err != nil {
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			continue
		}

		var data PlayingNowResponse
		if err := json.Unmarshal(body, &data); err != nil {
			continue
		}

		if len(data.Payload.Listens) > 0 {
			listening := data.Payload.Listens[0]
			track := listening.TrackMetadata
			data, _ := getData(db, username)
			
			if data == nil || *data != listening {
				sendPosts(track, sender)
				updateData(db, username, listening)
			}
		}
	}
}
