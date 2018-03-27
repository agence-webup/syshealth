package bolt

import (
	bolt "github.com/coreos/bbolt"
)

var openedDb *bolt.DB

// GetConnection returns an opened bolt instance
func GetConnection() (*bolt.DB, error) {
	if openedDb == nil {
		db, err := bolt.Open("syshealth.db", 0600, nil)
		if err != nil {
			return nil, err
		}
		openedDb = db
	}
	return openedDb, nil
}

// CloseConnection closes the opened connection, if any
func CloseConnection() error {
	if openedDb != nil {
		return openedDb.Close()
	}
	return nil
}
