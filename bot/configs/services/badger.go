package services

import (
	badger "github.com/dgraph-io/badger/v4"
)

var Badger *badger.DB

func NewBadger() (_ *badger.DB, err error) {
	Badger, err = badger.Open(badger.DefaultOptions("").WithInMemory(true))
	if err != nil {
		return nil, err
	}

	return Badger, err
}
