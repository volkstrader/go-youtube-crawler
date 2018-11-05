package main

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/buntdb"
	"github.com/tidwall/match"
	"google.golang.org/api/youtube/v3"
	"strings"
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

	//err = db.Update(func(tx *buntdb.Tx) error {
	//	tx.CreateIndex("title", "*", buntdb.IndexJSON("title"))
	//	tx.CreateIndex("desc", "*", buntdb.IndexJSON("description"))
	//	return nil
	//})
	//if err != nil {
	//	return database{}, fmt.Errorf("create JSON index failure, error: %s", err)
	//}

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

		key := getKey(video.Id, video.Snippet.Title)
		value := string(b)
		_, _, err = tx.Set(key, value, nil)

		log.Debug().
			Str("key", key).
			Str("title", video.Snippet.Title).
			//Str("value", value).
			Msg("save video to database")

		return err
	})
}

func getKey(id string, title string) string {
	return fmt.Sprintf("%s:title:%s", id, strings.ToLower(title))
}

type findResult struct {
	id      string
	title   string
	snippet youtube.VideoSnippet
}

// finding videos that match both title pattern and description pattern
// if input pattern is empty string, it is assumed as wildcard *
func (db database) find(titlePattern string, descPattern string) ([]findResult, error) {
	if titlePattern == "" {
		titlePattern = "*"
	}

	if descPattern == "" {
		descPattern = "*"
	}

	results := make([]findResult, 0, 50)

	err := db.store.View(func(tx *buntdb.Tx) error {
		pattern := fmt.Sprintf("*:title:%s", titlePattern)
		tx.AscendKeys(pattern, func(key, value string) bool {
			snippet := youtube.VideoSnippet{}
			json.Unmarshal([]byte(value), &snippet)

			if match.Match(snippet.Description, descPattern) {
				results = append(results, findResult{
					id:      key,
					title:   snippet.Title,
					snippet: snippet,
				})
			}

			return true
		})

		return nil
	})

	return results, err
}
