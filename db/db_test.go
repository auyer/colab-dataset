package db

import (
	"log"
	"os"
	"testing"
)

const (
	dbPath     = "./fastgate.db_test.go.db"
	testKey    = "TestKey"
	testValue  = 500
	testValue2 = -500
)

func TestDatabase(t *testing.T) {
	if _, err := os.Stat(dbPath); !os.IsNotExist(err) {
		err = os.RemoveAll(dbPath)
		if err != nil {
			log.Fatal("Unable to clean Test Database Before testing. Check for permissions.")
		}
	}
	datab, err := Init(dbPath)
	if err != nil {
		t.Errorf("Unable to Init Database")
	}
	err = InsertResource(testKey, testValue, datab)
	if err != nil {
		t.Errorf("Unable to Insert Tuple")
		t.FailNow()
	}
	value, err := GetResourceValue(testKey, datab)
	if err != nil {
		t.Errorf("Unable to Fetch Tuple")
		t.FailNow()

	}
	if value != testValue {
		t.Errorf("Received Value not mathing with what was inserted.")
		t.FailNow()

	}
	err = UpdateResource(testKey, testValue2, datab)
	if err != nil {
		t.Errorf("Unable to Update Tuple")
		t.FailNow()
	}
	value, err = GetResourceValue(testKey, datab)
	if err != nil {
		t.Errorf("Unable to Fetch Tuple after Updating")
		t.FailNow()

	}
	if value != testValue+testValue2 {
		t.Errorf("Received Value not mathing with what was inserted after updating.")
		t.FailNow()

	}
	err = datab.Close()
	if err != nil {
		t.Errorf("Failed at Closing Database")
		t.FailNow()

	}
	err = os.RemoveAll(dbPath)
	if err != nil {
		log.Printf("Unable to clean Test Database Aftere test. Check for permissions, and remove foleder '%s' or Future Tests might Fail", dbPath)
	}
}
