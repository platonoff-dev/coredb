package corekv

import "errors"

type DB struct {
}

func NewDB(path string) (*DB, error) {
	return nil, errors.New("not implemented")
}

func OpenDB(path string) (*DB, error) {
	return nil, errors.New("not implemented")
}

func (db *DB) Get(key []byte) ([]byte, error) {
	// Implement the logic to get a value by key from the database
	return nil, errors.New("not implemented")
}

func (db *DB) Put(key, value []byte) error {
	// Implement the logic to put a key-value pair into the database
	return errors.New("not implemented")
}

func (db *DB) Delete(key []byte) error {
	// Implement the logic to delete a key-value pair from the database
	return errors.New("not implemented")
}

// TODO: What to do if it's a large database?
// How to to scan only a part of the database? (table scan)
// Should we use a cursor or iterator?
// Interface not defined yet, so we can change it later
func (db *DB) Scan() ([][][]byte, error) {
	// Implement the logic to scan the database and return all key-value pairs
	return nil, errors.New("not implemented")
}

func (db *DB) Close() error {
	// Implement the logic to close the database connection
	return errors.New("not implemented")
}
