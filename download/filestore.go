package download

import (
	"io"
)

type FileStore interface {
	Delete(*Download) (bool, error)
	GetWriter(*Download) (io.WriteCloser, error)
	GetReader(*Download) (io.ReadCloser, error)
	Verify(*Download) (bool, error)
}
