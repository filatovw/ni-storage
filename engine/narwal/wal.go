package narwal

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/filatovw/ni-storage/engine"
	"github.com/filatovw/ni-storage/logger"
	"github.com/pkg/errors"
)

// Storage specific operations

type WAL struct {
	path          string
	rw            io.ReadWriteCloser
	lock          *sync.Mutex
	maxRecordSize int64
	log           logger.Logger
}

type action int

const (
	actionSet    action = 0
	actionDelete action = 1

	defaultMaxRecordSize = 2 << 24 // 16 MB
	defaultMaxBufferSize = 2 << 26 // 64 MB
)

type record struct {
	action   action
	len      int32
	data     []byte
	checksum uint32
}

func OpenWAL(path string, maxRecordSize int64, log logger.Logger) (*WAL, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, errors.Wrap(err, "path is not absolute")
	}
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, errors.Wrap(err, "create directory")
	}
	dataPath := filepath.Join(path, "data.wal")
	rw, err := os.OpenFile(dataPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return nil, errors.Wrap(err, "init storage")
	}
	return &WAL{
		path:          dataPath,
		rw:            rw,
		maxRecordSize: maxRecordSize,
		lock:          &sync.Mutex{},
		log:           log,
	}, nil
}

func (l *WAL) Close() error {
	return l.rw.Close()
}

func (l *WAL) Read() (map[string]engine.Record, error) {
	result := make(map[string]engine.Record)
	scan := bufio.NewScanner(l.rw)
	var e event

	scan.Split(bufio.ScanLines)
	for scan.Scan() {
		r := scan.Bytes()
		if err := json.Unmarshal(r, &e); err != nil {
			return nil, err
		}

		switch e.Action {
		case actionSet:
			result[e.Record.Key] = e.Record
		case actionDelete:
			delete(result, e.Record.Key)
		default:
			return nil, errors.New("unknown action")
		}
	}

	if err := scan.Err(); err != nil {
		return nil, errors.Wrap(err, "read error")
	}
	return result, nil
}

func (l *WAL) Write(e event) error {
	r, err := json.Marshal(e)
	if err != nil {
		return err
	}
	r = append(r, []byte("\n")...)
	if _, err = l.rw.Write(r); err != nil {
		return err
	}

	return nil
}

func (l WAL) Compact() error {
	return nil
}
