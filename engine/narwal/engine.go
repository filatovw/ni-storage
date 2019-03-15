package narwal

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"

	"github.com/filatovw/ni-storage/engine"
	"github.com/filatovw/ni-storage/logger"
	"github.com/pkg/errors"
)

type Narwal struct {
	log  logger.Logger
	data map[string]engine.Record // in memory data storage
	lock *sync.RWMutex
	wal  *WAL
}

type event struct {
	Record engine.Record `json:"record"`
	Action action        `json:"action"`
}

func (e *event) Bytes() []byte {
	return []byte(fmt.Sprintf("%d::%s", e.Action, e.Record))
}

func New(path string, log logger.Logger) (*Narwal, error) {
	wal, err := OpenWAL(path, DefaultMaxRecordSize)
	if err != nil {
		return nil, errors.Wrap(err, "open WAL")
	}
	snapshot, err := wal.Read()
	if err != nil {
		return nil, err
	}
	log.Infof("snapshot: %+v", snapshot)
	if snapshot == nil {
		snapshot = make(map[string]engine.Record)
	}
	return &Narwal{
		log:  log,
		wal:  wal,
		lock: &sync.RWMutex{},
		data: snapshot,
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
	s.set(record)
}

func (s *Narwal) Delete(key string) {
	s.lock.Lock()
	defer s.lock.Unlock()
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
	if err := s.wal.Write(event{Record: record, Action: actionSet}); err != nil {
		s.log.Error(err)
	}
	s.data[record.Key] = record
}

func (s *Narwal) delete(key string) {
	if err := s.wal.Write(event{Record: engine.Record{Key: key}, Action: actionDelete}); err != nil {
		s.log.Error(err)
	}
	delete(s.data, key)
}
