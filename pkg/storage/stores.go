package stores

import (
	"VoiceSculptor/pkg/util"
	"io"
	"net/http"
)

const (
	KindLocal = "local"
	KindOss   = "oss" // aliyun
	KindCos   = "cos" // tencent
)

var ErrInvalidPath = &util.Error{Code: http.StatusBadRequest, Message: "invalid path"}

var DefaultStoreKind = KindLocal

type Store interface {
	Read(key string) (io.ReadCloser, int64, error)
	Write(key string, r io.Reader) error
	Delete(key string) error
	Exists(key string) (bool, error)
	PublicURL(key string) string
}

func GetStore(kind string) Store {
	switch kind {
	case KindOss:
		return NewOssStore()
	case KindCos:
		return NewCosStore()
	default:
		return NewLocalStore()
	}
}

func Default() Store {
	return GetStore(DefaultStoreKind)
}
