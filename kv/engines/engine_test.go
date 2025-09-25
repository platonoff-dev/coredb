package engines

import (
	"testing"

	"github.com/platonoff-dev/coredb/kv/engines/eerrors"
	"github.com/platonoff-dev/coredb/kv/engines/mem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Engine interface {
	Get(key []byte) ([]byte, error)
	Put(key, value []byte) error
	Delete(key []byte) error
}

func setupEngines() map[string]Engine {
	return map[string]Engine{
		"mem": mem.NewMemEngine(),
	}
}

func TestEngines(t *testing.T) {
	cases := []struct {
		name     string
		testFunc func(Engine) error
	}{
		{
			name: "Get_success",
			testFunc: func(e Engine) error {
				require.NoError(t, e.Put([]byte("key"), []byte("value")))

				v, err := e.Get([]byte("key"))
				assert.NoError(t, err)
				assert.Equal(t, v, []byte("value"))

				return nil
			},
		},
		{
			name: "Get_not_found",
			testFunc: func(e Engine) error {
				v, err := e.Get([]byte("key"))
				assert.Error(t, err)
				assert.Equal(t, []byte(nil), v)
				assert.Equal(t, eerrors.ErrKeyNotFound, err)

				return nil
			},
		},
		{
			name: "PutNew_success",
			testFunc: func(e Engine) error {
				require.NoError(t, e.Put([]byte("key"), []byte("value")))

				v, err := e.Get([]byte("key"))
				assert.NoError(t, err)
				assert.Equal(t, v, []byte("value"))

				return nil
			},
		},
		{
			name: "PutExisting_success",
			testFunc: func(e Engine) error {
				require.NoError(t, e.Put([]byte("key"), []byte("value")))
				v, err := e.Get([]byte("key"))
				assert.NoError(t, err)
				assert.Equal(t, v, []byte("value"))

				require.NoError(t, e.Put([]byte("key"), []byte("value2")))
				v2, err := e.Get([]byte("key"))
				assert.NoError(t, err)
				assert.Equal(t, v2, []byte("value2"))

				return nil
			},
		},
		{
			name: "Delete_success",
			testFunc: func(e Engine) error {
				require.NoError(t, e.Put([]byte("key"), []byte("value")))
				v, err := e.Get([]byte("key"))
				assert.NoError(t, err)
				assert.Equal(t, v, []byte("value"))

				require.NoError(t, e.Delete([]byte("key")))
				v, err = e.Get([]byte("key"))
				assert.Error(t, err)
				assert.Equal(t, eerrors.ErrKeyNotFound, err)

				return nil
			},
		},
	}

	for _, tc := range cases {
		engines := setupEngines()
		for engineName, engine := range engines {
			t.Run(engineName+"_"+tc.name, func(t *testing.T) {
				assert.NoError(t, tc.testFunc(engine))
			})
		}
	}
}
