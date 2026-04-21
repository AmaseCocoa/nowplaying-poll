package main

import (
	"encoding/json"
	"fmt"
	"syscall"

	"golang.org/x/term"
	"go.etcd.io/bbolt"
)

func ScanSecret(ptr *string) error {
	byteSecret, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return err
	}

	*ptr = string(byteSecret)

	fmt.Println()
	return nil
}

func updateData(db *bbolt.DB, username string, listening PlayingNowPayload) {
	db.Update(func(tx *bbolt.Tx) error {
		bucket, _ := tx.CreateBucketIfNotExists([]byte("NowPlayingCache"))
		res, _ := json.Marshal(listening)
		return bucket.Put([]byte(username), res)
	})
}

func getData(db *bbolt.DB, username string) (*PlayingNowPayload, error) {
	var val []byte

	err := db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("NowPlayingCache"))
		if bucket == nil {
			return fmt.Errorf("bucket not found")
		}

		res := bucket.Get([]byte(username))
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
		var payload *PlayingNowPayload
		err := json.Unmarshal(val, &payload)
		if err != nil {
			return nil, err
		}
		return payload, nil
	}

	return nil, fmt.Errorf("failed to get cache")
}