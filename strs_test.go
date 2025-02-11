package rosedb

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestRoseDB_Set(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		testRoseDBSet(t, FileIO, KeyOnlyMemMode)
	})

	t.Run("mmap", func(t *testing.T) {
		testRoseDBSet(t, MMap, KeyOnlyMemMode)
	})

	t.Run("key-val-mem-mode", func(t *testing.T) {
		testRoseDBSet(t, FileIO, KeyValueMemMode)
	})
}

func TestRoseDB_Set_LogFileThreshold(t *testing.T) {
	path := filepath.Join("/tmp", "rosedb")
	opts := DefaultOptions(path)
	opts.IoType = MMap
	opts.LogFileSizeThreshold = 32 << 20
	db, err := Open(opts)
	assert.Nil(t, err)
	defer destroyDB(db)

	for i := 0; i < 600000; i++ {
		err := db.Set(GetKey(i), GetValue128B())
		assert.Nil(t, err)
	}
}

func TestRoseDB_Get(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		testRoseDBGet(t, FileIO, KeyOnlyMemMode)
	})

	t.Run("mmap", func(t *testing.T) {
		testRoseDBGet(t, MMap, KeyOnlyMemMode)
	})

	t.Run("key-val-mem-mode", func(t *testing.T) {
		testRoseDBGet(t, MMap, KeyValueMemMode)
	})
}

func TestRoseDB_Get_LogFileThreshold(t *testing.T) {
	path := filepath.Join("/tmp", "rosedb")
	opts := DefaultOptions(path)
	opts.IoType = MMap
	opts.LogFileSizeThreshold = 32 << 20
	db, err := Open(opts)
	assert.Nil(t, err)
	defer destroyDB(db)

	writeCount := 600000
	for i := 0; i <= writeCount; i++ {
		err := db.Set(GetKey(i), GetValue128B())
		assert.Nil(t, err)
	}

	rand.Seed(time.Now().Unix())
	for i := 0; i < 10000; i++ {
		key := GetKey(rand.Intn(writeCount))
		v, err := db.Get(key)
		assert.Nil(t, err)
		assert.NotNil(t, v)
	}
}

func testRoseDBSet(t *testing.T, ioType IOType, mode DataIndexMode) {
	path := filepath.Join("/tmp", "rosedb")
	opts := DefaultOptions(path)
	opts.IoType = ioType
	opts.IndexMode = mode
	db, err := Open(opts)
	assert.Nil(t, err)
	defer destroyDB(db)

	type args struct {
		key   []byte
		value []byte
	}
	tests := []struct {
		name    string
		db      *RoseDB
		args    args
		wantErr bool
	}{
		{
			"nil-key", db, args{key: nil, value: []byte("val-1")}, false,
		},
		{
			"nil-value", db, args{key: []byte("key-1"), value: nil}, false,
		},
		{
			"normal", db, args{key: []byte("key-1111"), value: []byte("value-1111")}, false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.db.Set(tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func testRoseDBGet(t *testing.T, ioType IOType, mode DataIndexMode) {
	path := filepath.Join("/tmp", "rosedb")
	opts := DefaultOptions(path)
	opts.IoType = ioType
	opts.IndexMode = mode
	db, err := Open(opts)
	assert.Nil(t, err)
	defer destroyDB(db)

	db.Set(nil, []byte("v-1111"))
	db.Set([]byte("k-1"), []byte("v-1"))
	db.Set([]byte("k-2"), []byte("v-2"))
	db.Set([]byte("k-3"), []byte("v-3"))
	db.Set([]byte("k-3"), []byte("v-333"))

	type args struct {
		key []byte
	}
	tests := []struct {
		name    string
		db      *RoseDB
		args    args
		want    []byte
		wantErr bool
	}{
		{
			"nil-key", db, args{key: nil}, nil, true,
		},
		{
			"normal", db, args{key: []byte("k-1")}, []byte("v-1"), false,
		},
		{
			"normal-rewrite", db, args{key: []byte("k-3")}, []byte("v-333"), false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.db.Get(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoseDB_Delete(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		testRoseDBDelete(t, FileIO, KeyOnlyMemMode)
	})

	t.Run("mmap", func(t *testing.T) {
		testRoseDBDelete(t, MMap, KeyOnlyMemMode)
	})

	t.Run("key-val-mem-mode", func(t *testing.T) {
		testRoseDBDelete(t, MMap, KeyValueMemMode)
	})
}

func TestRoseDB_Delete_MultiFiles(t *testing.T) {
	path := filepath.Join("/tmp", "rosedb")
	opts := DefaultOptions(path)
	opts.IoType = FileIO
	opts.LogFileSizeThreshold = 32 << 20
	db, err := Open(opts)
	assert.Nil(t, err)
	defer destroyDB(db)

	writeCount := 600000
	for i := 0; i <= writeCount; i++ {
		err := db.Set(GetKey(i), GetValue128B())
		assert.Nil(t, err)
	}

	var deletedKeys [][]byte
	rand.Seed(time.Now().Unix())
	for i := 0; i < 10000; i++ {
		key := GetKey(rand.Intn(writeCount))
		err := db.Delete(key)
		assert.Nil(t, err)
		deletedKeys = append(deletedKeys, key)
	}
	for _, k := range deletedKeys {
		_, err := db.Get(k)
		assert.Equal(t, ErrKeyNotFound, err)
	}
}

func testRoseDBDelete(t *testing.T, ioType IOType, mode DataIndexMode) {
	path := filepath.Join("/tmp", "rosedb")
	opts := DefaultOptions(path)
	opts.IoType = ioType
	opts.IndexMode = mode
	db, err := Open(opts)
	assert.Nil(t, err)
	defer destroyDB(db)

	db.Set(nil, []byte("v-1111"))
	db.Set([]byte("k-1"), []byte("v-1"))
	db.Set([]byte("k-3"), []byte("v-3"))
	db.Set([]byte("k-3"), []byte("v-333"))

	type args struct {
		key []byte
	}
	tests := []struct {
		name    string
		db      *RoseDB
		args    args
		wantErr bool
	}{
		{
			"nil", db, args{key: nil}, false,
		},
		{
			"normal-1", db, args{key: []byte("k-1")}, false,
		},
		{
			"normal-2", db, args{key: []byte("k-3")}, false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.db.Delete(tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRoseDB_SetEx(t *testing.T) {
	t.Run("key-only", func(t *testing.T) {
		testRoseDBSetEx(t, KeyOnlyMemMode)
	})

	t.Run("key-value", func(t *testing.T) {
		testRoseDBSetEx(t, KeyValueMemMode)
	})
}

func testRoseDBSetEx(t *testing.T, mode DataIndexMode) {
	path := filepath.Join("/tmp", "rosedb")
	opts := DefaultOptions(path)
	opts.IndexMode = mode
	db, err := Open(opts)
	assert.Nil(t, err)
	defer destroyDB(db)

	err = db.SetEx(GetKey(1), GetValue16B(), time.Millisecond*200)
	assert.Nil(t, err)
	time.Sleep(time.Millisecond * 205)
	v, err := db.Get(GetKey(1))
	assert.Equal(t, 0, len(v))
	assert.Equal(t, ErrKeyNotFound, err)

	err = db.SetEx(GetKey(2), GetValue16B(), time.Second*200)
	time.Sleep(time.Millisecond * 200)
	v1, err := db.Get(GetKey(2))
	assert.NotNil(t, v1)
	assert.Nil(t, err)

	// set a existed key.
	err = db.Set(GetKey(3), GetValue16B())
	assert.Nil(t, err)

	err = db.SetEx(GetKey(3), GetValue16B(), time.Millisecond*200)
	time.Sleep(time.Millisecond * 205)
	v2, err := db.Get(GetKey(3))
	assert.Equal(t, 0, len(v2))
	assert.Equal(t, ErrKeyNotFound, err)
}
