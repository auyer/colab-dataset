// Package db manages route storage for FastGate.
// The storage is performed by a Key-Value community database called Badger.
package db

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/dgraph-io/badger"
)

// DbPointer exported variable stores a pointer to the database initialized by the Init function.
var DbPointer *badger.DB
var DBSize int

// Init takes a path as input and reads / creates a bBadger database .
func Init(databasePath string) error {
	dbinfo := fmt.Sprintf(databasePath)

	var err error
	DbPointer, err = connectDB(dbinfo)
	return err

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

// GetDB provides a pointer to the database initialized by the Init function.
func GetDB() *badger.DB {
	return DbPointer
}

// InsertResource is a simple querry that inserts/updates the Resource tuple used by FastGate.
func InsertResource(key string, vote int) error {
	// Check if key existis
	_, err := GetResourceValue(key)
	if err == nil {
		return errors.New("Key Exists")
	}
	return DbPointer.Update(func(txn *badger.Txn) error {
		value := make([]byte, 4)
		binary.PutVarint(value, int64(vote))
		err := txn.Set([]byte(key), value)
		return err
	})
}

// UpdateResource is a simple querry that inserts/updates the Resource tuple used by FastGate.
func UpdateResource(key string, vote int) error {
	oldval, err := GetResourceValue(key)
	if err != nil {
		return err
	}
	return DbPointer.Update(func(txn *badger.Txn) error {
		value := make([]byte, 4)
		binary.PutVarint(value, int64(oldval+vote))
		err := txn.Set([]byte(key), value)
		return err
	})
}

// GetResource finds an address matching an key.
func GetResourceValue(key string) (value int, err error) {
	var result int64
	err = DbPointer.View(func(txn *badger.Txn) error {
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

func GetRandomKey() (value string, err error) {
	rcount := rand.New(rand.NewSource(time.Now().Unix())).Intn(DBSize)
	err = DbPointer.View(func(txn *badger.Txn) error {
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
				fmt.Printf("key=%s\n", k)
				return nil
			} else {
				acount += 1
			}

		}
		return nil
	})
	return
	// return , err
}
func CountDBSize() (value int) {
	_ = DbPointer.View(func(txn *badger.Txn) error {
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

func GetCurrentVotes() {
	_ = DbPointer.View(func(txn *badger.Txn) error {
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
			fmt.Printf("key=%s, value=%d\n", k, result)
		}
		return nil
	})
}
