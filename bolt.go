package core

import (
	"bytes"
	"sort"

	"github.com/boltdb/bolt"
)

// BoltDatabaseFactory creates Database objects for an underlying BoltDB instance.
type BoltDatabaseFactory struct {
	file string
}

// Connect returns the KeyValueDatabase instance for talking to a BoltDB file.
func (b BoltDatabaseFactory) Connect() (KeyValueDatabase, error) {
	return NewLeaf(b.file)
}

// NewLeaf creates a connection to a BoltDB file
func NewLeaf(file string) (KeyValueDatabase, error) {
	db, err := bolt.Open(file, 0600, nil)
	if err != nil {
		return nil, err
	}
	return &LeafDB{db}, nil
}

// LeafDB resembles a BoltDB connection
type LeafDB struct {
	db *bolt.DB
}

// GetOrCreateKeyspace returns a Keyspace implementation for the underlying BoltDB instance.
func (l *LeafDB) GetOrCreateKeyspace(name string) (ks Keyspace, err error) {
	err = l.db.Update(func(tx *bolt.Tx) error {
		_, er := tx.CreateBucketIfNotExists([]byte(name))

		ks = &BoltKeyspace{name, l.db}
		return er
	})
	return ks, err
}

func (l *LeafDB) Close() error {
	return l.db.Close()
}

func (l *LeafDB) DeleteKeyspace(name string) error {
	err := l.db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte(name))
	})
	return err
}

type BoltKeyspace struct {
	name string
	db   *bolt.DB
}

func (b *BoltKeyspace) GetName() string {
	return b.name
}

func (b *BoltKeyspace) List(keys []string, callback func([]byte, []byte)) error {
	// if no keys are searched for then return error
	if len(keys) == 0 {
		return ErrEmptyKeyList
	}

	// inplace lexigraphical sort
	sort.Strings(keys)

	// create lookup table
	lookup := make(map[string]bool)
	for _, k := range keys {
		lookup[k] = true
	}

	// create db view
	err := b.db.View(func(tx *bolt.Tx) error {

		// open bucket
		b := tx.Bucket([]byte(b.name))

		// create cursor
		c := b.Cursor()

		// iterate over bucket keys from first key to last
		last := []byte(keys[len(keys)-1])
		for k, v := c.Seek([]byte(keys[0])); k != nil && bytes.Compare(k, last) <= 0; k, v = c.Next() {

			// if key is what we are looking for
			if _, ok := lookup[string(k)]; ok {

				// call callback
				callback(k, v)
				// fmt.Printf("key=%s, value=%s\n", k, v)
			}
		}
		return nil
	})
	return err
}

func (b *BoltKeyspace) Insert(key string, value []byte) error {

	err := b.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(b.name))
		err := b.Put([]byte(key), value)
		return err
	})
	return err
}

func (b *BoltKeyspace) Get(key string) (value []byte, err error) {

	err = b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(b.name))
		value = b.Get([]byte(key))
		if value == nil {
			return ErrKeyNotFound
		}
		return nil
	})
	return
}

func (b *BoltKeyspace) Update(key string, value []byte) error {
	return b.Insert(key, value)
}

func (b *BoltKeyspace) Delete(key string) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(b.name))
		return b.Delete([]byte(key))
	})
}

func (b *BoltKeyspace) Size() (value int64) {
	b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(b.name))
		stats := bucket.Stats()
		value = int64(stats.KeyN)
		return nil
	})
	return
}

func (b *BoltKeyspace) ForEach(each ItemHandler) error {
	return b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(b.name))
		return b.ForEach(each)
	})
}

func (b *BoltKeyspace) Contains(key string) (exists bool, err error) {

	err = b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(b.name))
		value := b.Get([]byte(key))
		if value != nil {
			exists = true
		}
		return nil
	})

	return exists, err
}
