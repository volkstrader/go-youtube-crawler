package main

import (
	"fmt"
	"github.com/tidwall/buntdb"
	"testing"
)

func TestDatabase_Find(t *testing.T) {
	db, err := newDB("videos.db")
	if err != nil {
		t.Fatal(err)
	}

	db.store.View(func(tx *buntdb.Tx) error {
		idx, err := tx.Indexes()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(idx)
		return nil
	})

	results, err := db.find("*trump*", "*google*")
	if err != nil {
		t.Error(err)
	}

	if results == nil || len(results) == 0 {
		t.Fail()
	}
	t.Log(results)
}
