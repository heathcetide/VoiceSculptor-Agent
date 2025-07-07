package stores

import "io"

type OssStore struct {
}

// Delete implements Store.
func (o *OssStore) Delete(key string) error {
	panic("unimplemented")
}

// Exists implements Store.
func (o *OssStore) Exists(key string) (bool, error) {
	panic("unimplemented")
}

// Read implements Store.
func (o *OssStore) Read(key string) (io.ReadCloser, int64, error) {
	panic("unimplemented")
}

// Write implements Store.
func (o *OssStore) Write(key string, r io.Reader) error {
	panic("unimplemented")
}

func (o *OssStore) PublicURL(key string) string {
	panic("unimplemented")
}

func NewOssStore() Store {
	return &OssStore{}
}
