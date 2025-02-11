package rosedb

import (
	"bytes"
	"fmt"
	"github.com/flower-corp/rosedb/logger"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestOpen(t *testing.T) {
	path := filepath.Join("/tmp", "rosedb")
	t.Run("default", func(t *testing.T) {
		opts := DefaultOptions(path)
		db, err := Open(opts)
		defer destroyDB(db)
		assert.Nil(t, err)
		assert.NotNil(t, db)
	})

	t.Run("mmap", func(t *testing.T) {
		opts := DefaultOptions(path)
		opts.IoType = MMap
		db, err := Open(opts)
		defer destroyDB(db)
		assert.Nil(t, err)
		assert.NotNil(t, db)
	})
}

func TestLogFileGC(t *testing.T) {
	path := filepath.Join("/tmp", "rosedb")
	opts := DefaultOptions(path)
	opts.LogFileGCInterval = time.Second * 7
	opts.LogFileGCRatio = 0.00001
	db, err := Open(opts)
	defer destroyDB(db)
	if err != nil {
		t.Error("open db err ", err)
	}

	writeCount := 800000
	for i := 0; i < writeCount; i++ {
		err := db.Set(GetKey(i), GetValue128B())
		assert.Nil(t, err)
	}

	var deleted [][]byte
	rand.Seed(time.Now().Unix())
	for i := 0; i < 100000; i++ {
		k := rand.Intn(writeCount)
		key := GetKey(k)
		err := db.Delete(key)
		assert.Nil(t, err)
		deleted = append(deleted, key)
	}

	time.Sleep(time.Second * 12)
	for _, key := range deleted {
		_, err := db.Get(key)
		assert.Equal(t, err, ErrKeyNotFound)
	}
}

func TestInMemoryDataDump_List(t *testing.T) {
	path := filepath.Join("/tmp", "rosedb")
	opts := DefaultOptions(path)
	opts.InMemoryDataDumpInterval = time.Second * 3
	opts.LogFileSizeThreshold = 32 << 20
	db, err := Open(opts)
	defer destroyDB(db)
	if err != nil {
		t.Error("open db err ", err)
	}

	listKey := []byte("my_list")
	writeCount := 600000
	for i := 0; i < writeCount; i++ {
		v := GetValue128B()
		err := db.LPush(listKey, v)
		assert.Nil(t, err)
	}
	time.Sleep(time.Second * 6)

	lLen := db.LLen(listKey)
	assert.Equal(t, lLen, uint32(writeCount))
}

func TestInMemoryDataDump_Hash(t *testing.T) {
	path := filepath.Join("/tmp", "rosedb")
	opts := DefaultOptions(path)
	opts.InMemoryDataDumpInterval = time.Second * 3
	opts.LogFileSizeThreshold = 32 << 20
	db, err := Open(opts)
	defer destroyDB(db)
	assert.Nil(t, err)

	hashKey := []byte("my_hash")
	writeCount := 600000
	for i := 0; i < writeCount; i++ {
		err := db.HSet(hashKey, GetKey(i), GetValue128B())
		assert.Nil(t, err)
	}

	_ = db.Close()
	newdb, err := Open(opts)
	assert.Nil(t, err)
	assert.Equal(t, writeCount, newdb.HLen(hashKey))
}

func TestInMemoryDataDump_Set(t *testing.T) {
	path := filepath.Join("/tmp", "rosedb")
	opts := DefaultOptions(path)
	opts.InMemoryDataDumpInterval = time.Second * 3
	opts.LogFileSizeThreshold = 32 << 20
	db, err := Open(opts)
	defer destroyDB(db)
	assert.Nil(t, err)

	setKey := []byte("my_set")
	writeCount := 600000
	for i := 0; i < writeCount; i++ {
		err := db.SAdd(setKey, GetValue128B())
		assert.Nil(t, err)
	}

	card := db.SCard(setKey)
	assert.Equal(t, writeCount, card)
}

func TestInMemoryDataDump_ZSet(t *testing.T) {
	path := filepath.Join("/tmp", "rosedb")
	opts := DefaultOptions(path)
	opts.InMemoryDataDumpInterval = time.Second * 2
	opts.LogFileSizeThreshold = 32 << 20
	db, err := Open(opts)
	defer destroyDB(db)
	assert.Nil(t, err)

	zsetKey := []byte("my_zset")
	writeCount := 600000
	for i := 0; i < writeCount; i++ {
		err := db.ZAdd(zsetKey, float64(i*11), GetValue128B())
		assert.Nil(t, err)
	}

	card := db.ZCard(zsetKey)
	assert.Equal(t, writeCount, card)
}

func TestRoseDB_NewIterator(t *testing.T) {
	path := filepath.Join("/tmp", "rosedb")
	opts := DefaultOptions(path)
	db, err := Open(opts)
	defer destroyDB(db)
	if err != nil {
		t.Error("open db err ", err)
	}

	writeCount := 8
	for i := 0; i < writeCount; i++ {
		err := db.Set(GetKey(i), GetValue16B())
		assert.Nil(t, err)
	}

	iter := db.NewIterator(IteratorOptions{
		Limit: 20,
	})
	for iter.HasNext() {
		assert.NotNil(t, iter.Key())
		assert.NotNil(t, iter.Value())
	}
}

func destroyDB(db *RoseDB) {
	if db != nil {
		err := os.RemoveAll(db.opts.DBPath)
		if err != nil {
			logger.Errorf("destroy db err: %v", err)
		}
	}
}

const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"

func init() {
	rand.Seed(time.Now().Unix())
}

// GetKey length: 32 Bytes
func GetKey(n int) []byte {
	return []byte("kvstore-bench-key------" + fmt.Sprintf("%09d", n))
}

func GetValue16B() []byte {
	var str bytes.Buffer
	for i := 0; i < 16; i++ {
		str.WriteByte(alphabet[rand.Int()%36])
	}
	return []byte(str.String())
}

func GetValue128B() []byte {
	var str bytes.Buffer
	for i := 0; i < 128; i++ {
		str.WriteByte(alphabet[rand.Int()%36])
	}
	return []byte(str.String())
}

func GetValue4K() []byte {
	var str bytes.Buffer
	for i := 0; i < 4096; i++ {
		str.WriteByte(alphabet[rand.Int()%36])
	}
	return []byte(str.String())
}
