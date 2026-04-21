package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mattn/go-mastodon"
	"go.etcd.io/bbolt"
)

const bucketName = "MastodonCredentials"

type MastodonCredential struct {
	AccessToken string `json:"access_token"`
	ClientID string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type MastodonSender struct {
	host string
	db *bbolt.DB
}

func (r MastodonSender) getTokenFromCache(host string) (*MastodonCredential, error) {
	var val []byte

	err := r.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return fmt.Errorf("bucket not found")
		}

		res := bucket.Get([]byte(host))
		if res != nil {
			val = make([]byte, len(res))
			copy(val, res)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if val != nil {
		var payload *MastodonCredential
		err := json.Unmarshal(val, &payload)
		if err != nil {
			return nil, err
		}
		return payload, nil
	}

	return nil, fmt.Errorf("failed to get cache")
}

func (r MastodonSender) GetToken() *mastodon.Client {
	token, _ := r.getTokenFromCache(r.host)
	
	if token != nil {
		return mastodon.NewClient(&mastodon.Config{
			Server:       r.host,
			ClientID:     token.ClientID,
			AccessToken: token.AccessToken,
			ClientSecret: token.ClientSecret,
		})
	}
	
	client := mastodon.NewClient(&mastodon.Config{
		Server:       r.host,
	})
	app, err := mastodon.RegisterApp(context.Background(), &mastodon.AppConfig{
	    Server:     r.host,
	    ClientName: "NowPlaying AutoPost",
	    Scopes:     "write:statuses",
	    Website:    "https://github.com/AmaseCocoa/nowplaying-poll",
	    RedirectURIs: "urn:ietf:wg:oauth:2.0:oob",
	})
	
	if err != nil {
    	log.Fatalf("Failed to register application: %v", err)
	}
	
	fmt.Printf("1. Please Authenticate in this url:\n%s\n\n", app.AuthURI)
	fmt.Print("2. Enter Token: ")
	var code string
	ScanSecret(&code)
	
	client = mastodon.NewClient(&mastodon.Config{
		Server:       r.host,
		ClientID:     app.ClientID,
		ClientSecret: app.ClientSecret,
	})

	err = client.GetUserAccessToken(context.Background(), code, "urn:ietf:wg:oauth:2.0:oob")
	if err != nil {
		log.Fatal(err)
	}
	
	r.db.Update(func(tx *bbolt.Tx) error {
		bucket, _ := tx.CreateBucketIfNotExists([]byte(bucketName))
		res, _ := json.Marshal(MastodonCredential {
			AccessToken: client.Config.AccessToken,
			ClientID: app.ClientID,
			ClientSecret: app.ClientSecret,
		})
		return bucket.Put([]byte(r.host), res)
	})
	
	return client
}