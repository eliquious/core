package core

import (
	// "encoding/json"

	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	// "time"
)

func TestLeafDBNew(t *testing.T) {
	file := os.TempDir() + "/leaf.db"

	leaf, err := NewLeaf(file)
	defer leaf.Close()

	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, leaf)
}

func TestLeafCreateKeyspace(t *testing.T) {
	file := os.TempDir() + "/leaf.db"

	db := BoltDatabaseFactory{file}
	leaf, err := db.Connect()
	defer leaf.Close()

	// created db
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, leaf)

	// create keyspace
	ks, err := leaf.GetOrCreateKeyspace("users")
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, ks)

	// keyspace name
	assert.Equal(t, "users", ks.GetName())

	// cleanup ks
	leaf.DeleteKeyspace("users")
}

func TestBoltKeyspaceInsert(t *testing.T) {
	file := os.TempDir() + "/leaf.db"

	db := BoltDatabaseFactory{file}
	leaf, err := db.Connect()
	defer leaf.Close()

	// created db
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, leaf)

	// create keyspace
	ks, err := leaf.GetOrCreateKeyspace("users")
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, ks)

	// keyspace name
	assert.Equal(t, "users", ks.GetName())

	// insert
	err = ks.Insert("username", []byte{})
	assert.Equal(t, nil, err)

	// cleanup ks
	leaf.DeleteKeyspace("users")
}

func TestBoltKeyspaceList(t *testing.T) {
	file := os.TempDir() + "/leaf.db"

	db := BoltDatabaseFactory{file}
	leaf, err := db.Connect()
	defer leaf.Close()

	// created db
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, leaf)

	// create keyspace
	ks, err := leaf.GetOrCreateKeyspace("users")
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, ks)

	// keyspace name
	assert.Equal(t, "users", ks.GetName())

	// insert
	lookup := make(map[string]bool)
	err = ks.Insert("user1", []byte("1"))
	assert.Equal(t, nil, err)
	lookup["user1"] = false

	err = ks.Insert("user2", []byte("2"))
	assert.Equal(t, nil, err)
	lookup["user2"] = false

	err = ks.Insert("user3", []byte("3"))
	assert.Equal(t, nil, err)
	lookup["user3"] = false

	err = ks.Insert("user4", []byte("4"))
	assert.Equal(t, nil, err)
	lookup["user4"] = false

	err = ks.Insert("user5", []byte("5"))
	assert.Equal(t, nil, err)
	lookup["user5"] = false

	err = ks.Insert("user6", []byte("6"))
	assert.Equal(t, nil, err)
	lookup["user6"] = false

	// list keys
	// start := time.Now()
	err = ks.List([]string{"user2", "user4", "user5"}, func(k, v []byte) {
		lookup[string(k)] = true
	})
	// fmt.Println(time.Now().Sub(start))

	// test found
	assert.False(t, lookup["user1"])
	assert.True(t, lookup["user2"])
	assert.False(t, lookup["user3"])
	assert.True(t, lookup["user4"])
	assert.True(t, lookup["user5"])
	assert.False(t, lookup["user6"])

	// cleanup ks
	leaf.DeleteKeyspace("users")
}

func TestBoltKeyspaceGet(t *testing.T) {
	file := os.TempDir() + "/leaf.db"

	db := BoltDatabaseFactory{file}
	leaf, err := db.Connect()
	defer leaf.Close()

	// created db
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, leaf)

	// create keyspace
	ks, err := leaf.GetOrCreateKeyspace("users")
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, ks)

	// keyspace name
	assert.Equal(t, "users", ks.GetName())

	// insert
	err = ks.Insert("user1", []byte("1"))
	assert.Equal(t, nil, err)

	// get values
	value, err := ks.Get("user1")
	assert.Nil(t, err)
	assert.NotNil(t, value)

	value, err = ks.Get("user2")
	assert.Equal(t, ErrKeyNotFound, err)
	assert.Nil(t, value)
	// fmt.Println(string(value))

	leaf.DeleteKeyspace("users")
}

func TestBoltKeyspaceUpdate(t *testing.T) {
	file := os.TempDir() + "/leaf.db"

	db := BoltDatabaseFactory{file}
	leaf, err := db.Connect()
	defer leaf.Close()

	// created db
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, leaf)

	// create keyspace
	ks, err := leaf.GetOrCreateKeyspace("users")
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, ks)

	// keyspace name
	assert.Equal(t, "users", ks.GetName())

	// insert
	err = ks.Insert("user1", []byte("1"))
	assert.Equal(t, nil, err)

	// get values
	value, err := ks.Get("user1")
	assert.Nil(t, err)
	assert.Equal(t, []byte("1"), value)

	// update
	err = ks.Update("user1", []byte("2"))
	assert.Nil(t, err)

	// get new value
	value, err = ks.Get("user1")
	assert.Nil(t, err)
	assert.Equal(t, []byte("2"), value)

	leaf.DeleteKeyspace("users")
}

func TestBoltKeyspaceDelete(t *testing.T) {
	file := os.TempDir() + "/leaf.db"

	db := BoltDatabaseFactory{file}
	leaf, err := db.Connect()
	defer leaf.Close()

	// created db
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, leaf)

	// create keyspace
	ks, err := leaf.GetOrCreateKeyspace("users")
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, ks)

	// keyspace name
	assert.Equal(t, "users", ks.GetName())

	// insert
	err = ks.Insert("user1", []byte("1"))
	assert.Equal(t, nil, err)

	// get values
	value, err := ks.Get("user1")
	assert.Nil(t, err)
	assert.Equal(t, []byte("1"), value)

	// delete
	err = ks.Delete("user1")
	assert.Nil(t, err)

	// get new value
	value, err = ks.Get("user1")
	assert.Nil(t, value)
	assert.Equal(t, ErrKeyNotFound, err)

	leaf.DeleteKeyspace("users")
}

func TestBoltKeyspaceSize(t *testing.T) {
	file := os.TempDir() + "/leaf.db"

	db := BoltDatabaseFactory{file}
	leaf, err := db.Connect()
	defer leaf.Close()

	// created db
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, leaf)

	// create keyspace
	ks, err := leaf.GetOrCreateKeyspace("users")
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, ks)

	// keyspace size
	assert.Equal(t, int64(0), ks.Size())

	// insert
	err = ks.Insert("user1", []byte("1"))
	assert.Equal(t, nil, err)
	assert.Equal(t, int64(1), ks.Size())

	err = ks.Insert("user2", []byte("2"))
	assert.Equal(t, nil, err)
	assert.Equal(t, int64(2), ks.Size())

	leaf.DeleteKeyspace("users")
}

func TestBoltKeyspaceForEach(t *testing.T) {
	file := os.TempDir() + "/leaf.db"

	db := BoltDatabaseFactory{file}
	leaf, err := db.Connect()
	defer leaf.Close()

	// created db
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, leaf)

	// create keyspace
	ks, err := leaf.GetOrCreateKeyspace("users")
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, ks)

	// insert
	items := make(map[string]bool)

	err = ks.Insert("user1", []byte("1"))
	assert.Equal(t, nil, err)

	err = ks.Insert("user2", []byte("2"))
	assert.Equal(t, nil, err)

	// iterate
	ks.ForEach(func(k, v []byte) error {
		items[string(k)] = true
		return nil
	})

	// test existence
	assert.Equal(t, true, items["user1"])
	assert.Equal(t, true, items["user2"])

	leaf.DeleteKeyspace("users")
}

func TestBoltKeyspaceContains(t *testing.T) {
	file := os.TempDir() + "/leaf.db"

	db := BoltDatabaseFactory{file}
	leaf, err := db.Connect()
	defer leaf.Close()

	// created db
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, leaf)

	// create keyspace
	ks, err := leaf.GetOrCreateKeyspace("users")
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, ks)

	// insert
	err = ks.Insert("user1", []byte("1"))
	assert.Equal(t, nil, err)

	// test existence
	exists, err := ks.Contains("user1")
	assert.Equal(t, true, exists)

	exists, err = ks.Contains("user2")
	assert.Equal(t, false, exists)

	leaf.DeleteKeyspace("users")
}
