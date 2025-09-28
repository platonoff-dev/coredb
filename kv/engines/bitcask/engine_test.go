package bitcask

import (
	"testing"

	"github.com/platonoff-dev/coredb/kv/engines/common_errors"
	"github.com/stretchr/testify/assert"
)

func setupTestEngine(t *testing.T) *Engine {
	t.Helper()

	dir := t.TempDir()
	engine := New(dir)
	if err := engine.Open(); err != nil {
		t.Fatal(err)
	}

	return engine
}

func TestBasicFlow(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	_, err := engine.Get([]byte("test_key"))
	assert.Error(t, err)
	assert.Equal(t, err, common_errors.ErrKeyNotFound)

	key := []byte("test_key")
	value := []byte("test_value")

	err = engine.Put(key, value)
	assert.NoError(t, err)
	resultValue, err := engine.Get(key)
	assert.NoError(t, err)
	assert.Equal(t, value, resultValue)

	newValue := []byte("new_value")
	err = engine.Put(key, newValue)
	assert.NoError(t, err)
	resultValue, err = engine.Get(key)
	assert.NoError(t, err)
	assert.Equal(t, newValue, resultValue)
}
