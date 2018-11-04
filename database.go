package main

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/buntdb"
	"google.golang.org/api/youtube/v3"
)

type database struct {
	store *buntdb.DB
}

func newDB(dbfile string) (database, error) {
	if dbfile == "" {
		dbfile = ":memory:"
	}

	db, err := buntdb.Open(dbfile)
	if err != nil {
		return database{}, err
	}

	err = db.Update(func(tx *buntdb.Tx) error {
		tx.CreateIndex("title", "*", buntdb.IndexJSON("title"))
		tx.CreateIndex("desc", "*", buntdb.IndexJSON("description"))
		return nil
	})
	if err != nil {
		return database{}, fmt.Errorf("create JSON index failure, error: %s", err)
	}

	return database{
		store: db,
	}, nil
}

func (db database) close() {
	db.store.Close()
}

func (db database) save(video *youtube.Video) error {
	return db.store.Update(func(tx *buntdb.Tx) error {
		b, err := json.Marshal(video.Snippet)
		if err != nil {
			return fmt.Errorf("JSON marshal failure: error: %s", err)
		}

		value := string(b)
		_, _, err = tx.Set(video.Id, value, nil)

		log.Debug().
			Str("key", video.Id).
			Str("title", video.Snippet.Title).
			//Str("value", value).
			Msg("save video to database")

		return err
	})
}

type findResult struct {
	id          string
	title       string
	description string
	JSON        json.RawMessage
}

//func (db database) find(titlePattern string, descPattern string) ([]findResult, error) {
//
//}
