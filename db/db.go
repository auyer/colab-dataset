// Package db manages voting storage for.
// The storage is performed by a Key-Value community database called Badger.
package db

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/dgraph-io/badger"
)

// Vote strcucture used to process voting from the API with strings instead of numbers. This is intended to make voting safer, and avoid requests with a value bigger than 1.
type Vote struct {
	Key  string `json:"Key"`
	Vote string `json:"Vote"`
}

// VoteInt structure used to process votes inside the API, including sorting.
type VoteInt struct {
	Key  string `json:"Key"`
	Vote int    `json:"Vote"`
}

// VoteIntAmt structure is used to merge data from both databases (Votes and Amount of Votes)
type VoteIntAmt struct {
	Key        string `json:"Key"`
	Vote       int    `json:"Vote"`
	TotalVotes int    `json"TotalVotes"`
}

// Init takes a path as input and reads / creates a bBadger database .
func Init(databasePath string) (*badger.DB, error) {
	dbinfo := fmt.Sprintf(databasePath)
	return connectDB(dbinfo)
}

// connectDB manages the database connection and configuration.
func connectDB(databasePath string) (*badger.DB, error) {

	opts := badger.DefaultOptions
	opts.Dir = databasePath
	opts.ValueDir = databasePath
	db, err := badger.Open(opts)

	if err != nil {
		return nil, err
	}
	return db, nil
}

// InsertResource is a simple querry that inserts/updates the Resource tuple used by .
func InsertResource(key string, vote int, dbpointer *badger.DB) error {
	// Check if key existis
	_, err := GetResourceValue(key, dbpointer)
	if err == nil {
		return errors.New("Key Exists")
	}
	return dbpointer.Update(func(txn *badger.Txn) error {
		value := make([]byte, 4)
		binary.PutVarint(value, int64(vote))
		err := txn.Set([]byte(key), value)
		return err
	})
}

// UpdateResource is a simple querry that inserts/updates the Resource tuple.
func UpdateResource(key string, vote int, dbpointer *badger.DB) error {
	oldval, err := GetResourceValue(key, dbpointer)
	if err != nil {
		return err
	}
	return dbpointer.Update(func(txn *badger.Txn) error {
		value := make([]byte, 4)
		binary.PutVarint(value, int64(oldval+vote))
		err := txn.Set([]byte(key), value)
		return err
	})
}

// GetResourceValue returns the value of a specified resource
func GetResourceValue(key string, dbpointer *badger.DB) (value int, err error) {
	var result int64
	err = dbpointer.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		val, err := item.Value()
		if err != nil {
			return err
		}
		var berr int
		result, berr = binary.Varint(val)
		if berr == 0 {
			return errors.New("Failed to Read Binary")
		}
		//copy(result, val)
		return err
	})

	return int(result), err
}

// GetRandomKey returns a random key from the database
func GetRandomKey(dbpointer *badger.DB, dbsize int) (value string, err error) {
	rcount := rand.New(rand.NewSource(time.Now().Unix())).Intn(dbsize)
	err = dbpointer.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		acount := 0
		for it.Rewind(); it.Valid(); it.Next() {
			if acount == rcount {
				item := it.Item()
				k := item.Key()
				value = string(k)
				return nil
			} else {
				acount += 1
			}
		}
		return nil
	})
	return
}

// GetSortedKey will return the key with the smallest value
func GetSortedKey(dbpointer *badger.DB) (topKey string, err error) {
	var list []VoteInt
	err = dbpointer.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			v, err := item.Value()
			if err != nil {
				return err
			}
			result, _ := binary.Varint(v)
			list = append(list, VoteInt{string(k), int(result)})
		}
		return nil
	})
	fmt.Print("HERE!")
	sort.Slice(list, func(i, j int) bool {
		return list[i].Vote < list[j].Vote
	})
	topKey = list[0].Key
	return
}

// CountDBSize will return the amount of entries in the database
func CountDBSize(dbpointer *badger.DB) (value int) {
	_ = dbpointer.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			// item := it.Item()
			value += 1
		}
		return nil
	})
	return
}

// GetCurrentVotes will return a list of the votes in the database.
func GetCurrentVotes(dbpointer *badger.DB) (list []VoteInt, err error) {
	// list := make(chan struct {string; string})
	err = dbpointer.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			v, err := item.Value()
			if err != nil {
				return err
			}
			result, _ := binary.Varint(v)
			list = append(list, VoteInt{string(k), int(result)})
		}
		return nil
	})
	return
}
