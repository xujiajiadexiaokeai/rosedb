package hash

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var key = "my_hash"

func InitHash() *Hash {
	hash := New()

	hash.HSet(key, "a", []byte("hash_data_001"))
	hash.HSet(key, "b", []byte("hash_data_002"))
	hash.HSet(key, "c", []byte("hash_data_003"))

	return hash
}

func TestNew(t *testing.T) {
	hash := New()
	assert.NotEqual(t, hash, nil)
}

func TestHash_HSet(t *testing.T) {
	hash := InitHash()
	r1 := hash.HSet("my_hash", "d", []byte("123"))
	assert.Equal(t, r1, 1)
	r2 := hash.HSet("my_hash", "d", []byte("123"))
	assert.Equal(t, r2, 0)
	r3 := hash.HSet("my_hash", "e", []byte("234"))
	assert.Equal(t, r3, 1)
}

func TestHash_HSetNx(t *testing.T) {
	hash := InitHash()
	r1 := hash.HSetNx(key, "a", []byte("new one"))
	assert.Equal(t, r1, 0)
	r2 := hash.HSetNx(key, "f", []byte("d-new one"))
	assert.Equal(t, r2, 1)
	r3 := hash.HSetNx(key, "f", []byte("d-new one"))
	assert.Equal(t, r3, 0)
}

func TestHash_HGet(t *testing.T) {
	hash := InitHash()

	val := hash.HGet(key, "a")
	assert.Equal(t, []byte("hash_data_001"), val)
	valNotExist := hash.HGet(key, "m")
	assert.Equal(t, []byte(nil), valNotExist)
}

func TestHash_HGetAll(t *testing.T) {
	hash := InitHash()

	vals := hash.HGetAll(key)
	assert.Equal(t, 6, len(vals))
}

func TestHash_HDel(t *testing.T) {
	hash := InitHash()
	//delete existed filed,return 1
	v0 := hash.HDel(key, "a")
	assert.Equal(t, true, v0)
	//delete same field twice,return 0
	v1 := hash.HDel(key, "a")
	assert.Equal(t, false, v1)
	//delete non existing field,expect 0
	v2 := hash.HDel(key, "m")
	assert.Equal(t, false, v2)
}

func TestHash_HExists(t *testing.T) {
	hash := InitHash()
	// key and field both exist
	exist := hash.HExists(key, "a")
	assert.Equal(t, true, exist)
	// key is non existing
	keyNot := hash.HExists("non exiting key", "a")
	assert.Equal(t, false, keyNot)
	not := hash.HExists(key, "m")
	assert.Equal(t, false, not)

}

func TestHash_HKeys(t *testing.T) {
	hash := InitHash()

	keys := hash.HKeys(key)
	assert.Equal(t, len(keys), 3)
	res := hash.HKeys("no")
	assert.Equal(t, len(res), 0)
}

func TestHash_HVals(t *testing.T) {
	hash := InitHash()
	values := hash.HVals(key)
	for _, v := range values {
		assert.NotNil(t, v)
	}
}

func TestHash_HLen(t *testing.T) {
	hash := InitHash()
	assert.Equal(t, 3, hash.HLen(key))
}

func TestHash_HClear(t *testing.T) {
	hash := InitHash()
	hash.HClear(key)

	v := hash.HGet(key, "a")
	assert.Equal(t, len(v), 0)
}

func TestHash_HKeyExists(t *testing.T) {
	hash := InitHash()
	exists := hash.HKeyExists(key)
	assert.Equal(t, exists, true)

	hash.HClear(key)

	exists1 := hash.HKeyExists(key)
	assert.Equal(t, exists1, false)
}
