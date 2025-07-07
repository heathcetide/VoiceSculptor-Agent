package stores

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCosStore(t *testing.T) {

	store := NewCosStore().(*CosStore)
	assert.NotNil(t, store)
	fname := "test.txt"
	ok, err := store.Exists(fname)
	assert.NoError(t, err)
	assert.False(t, ok)

	err = store.Write(fname, bytes.NewReader([]byte("hello")))
	assert.NoError(t, err)

	r, size, err := store.Read(fname)
	assert.Equal(t, int64(200048), size)
	assert.NoError(t, err)
	fmt.Println(r)

	err = store.Delete(fname)
	assert.NoError(t, err)
}
