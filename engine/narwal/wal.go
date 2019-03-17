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

type action int

const (
	actionSet    action = 0
	actionDelete action = 1

	defaultMaxRecordSize = 2 << 24 // 16 MB
)

// WAL log-file in append mode
type WAL struct {
	maxRecordSize int
	path          string
	rw            io.ReadWriteCloser
	lock          *sync.Mutex
	log           logger.Logger
}

// OpenWAL open log or create it if it doesn't exist
func OpenWAL(log logger.Logger, path string, maxRecordSize int) (*WAL, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, errors.Wrap(err, "path is not absolute")
	}
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, errors.Wrap(err, "create directory")
	}
	dataPath := filepath.Join(path, "narwal.wal")
	rw, err := os.OpenFile(dataPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return nil, errors.Wrap(err, "init storage")
	}
	return &WAL{
		maxRecordSize: maxRecordSize,
		path:          dataPath,
		rw:            rw,
		lock:          &sync.Mutex{},
		log:           log,
	}, nil
}

// Close log
func (l *WAL) Close() error {
	l.lock.Lock()
	defer l.lock.Unlock()
	return l.rw.Close()
}

// Read snapshot from log-file
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

// Write event into log-file
func (l *WAL) Write(e event) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	if len(e.Record.Value) > l.maxRecordSize {
		return errors.New("entity is too large")
	}

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
