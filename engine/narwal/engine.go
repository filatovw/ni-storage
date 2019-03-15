package narwal

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/filatovw/ni-storage/engine"
	"github.com/filatovw/ni-storage/logger"
	"github.com/pkg/errors"
)

type action int

const (
	Set    action = 0
	Delete action = 1
)

type event struct {
	record engine.Record
	action action
}

type Narwal struct {
	log  logger.Logger
	data map[string]engine.Record // in memory data storage
	lock *sync.RWMutex
	w    io.WriteCloser // persistent data storage
	// queue chan event     // internal queue
}

func New(path string, log logger.Logger) (*Narwal, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, errors.Wrap(err, "path is not absolute")
	}
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, errors.Wrap(err, "create directory")
	}

	w, err := os.OpenFile(filepath.Join(path, "data.wal"), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return nil, errors.Wrap(err, "init storage")
	}
	return &Narwal{
		log:  log,
		w:    w,
		lock: &sync.RWMutex{},
		data: make(map[string]engine.Record),
		//	queue: make(chan event)
	}, nil
}

func (s *Narwal) Exists(key string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	_, ok := s.data[key]
	return ok
}

func (s *Narwal) Get(key string) (engine.Record, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	record, ok := s.data[key]
	if !ok {
		return engine.Null, false
	}
	return record, true
}

func (s *Narwal) Set(record engine.Record) {
	s.lock.Lock()
	defer s.lock.Unlock()
	/*
		s.queue <- event{
			record: record,
			action: Set,
		}
	*/
	s.set(record)
}

func (s *Narwal) Delete(key string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	/*
		s.queue <- event{
			record: engine.Record{Key: key},
			action: Delete,
		}
	*/
	s.delete(key)
}

func (s *Narwal) Filter(pattern string) ([]engine.Record, error) {
	regPattern := strings.ReplaceAll(pattern, "$", ".*")
	log.Printf("pattern %s", regPattern)
	exp, err := regexp.Compile(regPattern)
	if err != nil {
		return nil, err
	}

	s.lock.RLock()
	defer s.lock.RUnlock()

	results := []engine.Record{}
	for _, v := range s.data {
		if exp.MatchString(v.Value) {
			results = append(results, v)
		}
	}
	return results, nil
}

func (s *Narwal) GetAll() ([]engine.Record, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	results := []engine.Record{}
	for _, v := range s.data {
		results = append(results, v)
	}
	return results, nil
}

func (s *Narwal) DeleteAll() {
	s.lock.Lock()
	defer s.lock.Unlock()
	for k := range s.data {
		s.delete(k)
	}
}

func (s *Narwal) set(record engine.Record) {
	s.data[record.Key] = record
}

func (s *Narwal) delete(key string) {
	delete(s.data, key)
}
