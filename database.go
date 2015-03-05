package core

import "errors"

type KeyValuePair struct {
	Key   []byte
	Value []byte
}

type ItemHandler func(k, v []byte) error

var (
	// ErrKeyNotFound is returned if a Keyspace did not contain the key
	ErrKeyNotFound = errors.New("Key does not exist")

	// ErrEmptyKeyList is returned if Keyspace.List() is called with no keys
	ErrEmptyKeyList = errors.New("Empty key list")
)

// DatabaseConnectionFactory creates Database instances when Connect() is called
type DatabaseConnectionFactory interface {
	Connect() (KeyValueDatabase, error)
}

// Keyspace is an interface for Database keyspaces. It is used as a wrapper
// for database actions.
type Keyspace interface {
	GetName() string
	List([]string, func([]byte, []byte)) error
	Insert(string, []byte) error
	Get(string) ([]byte, error)
	Update(string, []byte) error
	Delete(string) error
	Size() int64
	ForEach(ItemHandler) error
	Contains(string) (bool, error)
}

// KeyValueDatabase is used as an interface for multiple backends and wraps any specific implementations.
type KeyValueDatabase interface {
	GetOrCreateKeyspace(string) (Keyspace, error)
	DeleteKeyspace(string) error
	Close() error
}
