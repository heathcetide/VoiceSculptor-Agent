package stores

import (
	"VoiceSculptor/pkg/util"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var UploadDir string = "/tmp"
var MediaPrefix string = "/media"

type LocalStore struct {
	Root       string
	NewDirPerm os.FileMode
}

// Delete implements Store.
func (l *LocalStore) Delete(key string) error {
	fname := filepath.Clean(filepath.Join(l.Root, key))
	if !strings.HasPrefix(fname, l.Root) {
		return ErrInvalidPath
	}
	return os.Remove(fname)
}

// Exists implements Store.
func (l *LocalStore) Exists(key string) (bool, error) {
	fname := filepath.Clean(filepath.Join(l.Root, key))
	if !strings.HasPrefix(fname, l.Root) {
		return false, ErrInvalidPath
	}
	_, err := os.Stat(fname)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Read implements Store.
func (l *LocalStore) Read(key string) (io.ReadCloser, int64, error) {
	fname := filepath.Clean(filepath.Join(l.Root, key))
	if !strings.HasPrefix(fname, l.Root) {
		return nil, 0, ErrInvalidPath
	}
	st, err := os.Stat(fname)
	if err != nil {
		return nil, 0, err
	}
	f, err := os.Open(fname)
	if err != nil {
		return nil, 0, err
	}
	return f, st.Size(), nil
}

// Write implements Store.
func (l *LocalStore) Write(key string, r io.Reader) error {
	fname := filepath.Clean(filepath.Join(l.Root, key))
	if !strings.HasPrefix(fname, l.Root) {
		return ErrInvalidPath
	}
	dir := filepath.Dir(fname)
	err := os.MkdirAll(dir, l.NewDirPerm)
	if err != nil {
		return err
	}
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}

func (l *LocalStore) PublicURL(key string) string {
	return filepath.Join(util.GetEnv("MEDIA_PREFIX"), key)
}

func NewLocalStore() Store {
	s := &LocalStore{
		Root:       util.GetEnv("UPLOAD_DIR"),
		NewDirPerm: 0755,
	}
	return s
}
