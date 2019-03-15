package narwal

import (
	"io"
	"os"
)

// Storage specific operations

type WAL struct {
	w io.WriteCloser
}

func Open(path string) (*WAL, error) {
	f, err := os.OpenFile(path, os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}
	return &WAL{w: f}, nil
}

func (l WAL) Read() error {
	return nil
}

func (l WAL) Write(e event) error {
	return nil
}

func (l WAL) Compact() error {
	return nil
}
