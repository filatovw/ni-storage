package narwal

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/filatovw/ni-storage/engine"
	"github.com/filatovw/ni-storage/logger"
	"github.com/pkg/errors"
)

type Narwal struct {
	log    logger.Logger
	lock   *sync.RWMutex
	data   map[string]engine.Record
	wal    *WAL
	ttlIdx *TTLIdx
}

type TTLIdx struct {
	stack []ttlRecord
}

type ttlRecord struct {
	until time.Time
	key   string
}

func (idx *TTLIdx) Push(r ttlRecord) {
	if len(idx.stack) == 0 {
		idx.stack = append(idx.stack, r)
		return
	}

	i := 0
	for _, e := range idx.stack {
		if r.until.After(e.until) {
			i++
			break
		}
	}
	idx.stack = append(idx.stack[:i], append([]ttlRecord{r}, idx.stack[i:]...)...)
}

func (idx *TTLIdx) Delete(k string) {
	i := 0
	for _, e := range idx.stack {
		if e.key == k {
			break
		}
		i++
	}
	copy(idx.stack[i:], idx.stack[i+1:])
	idx.stack = idx.stack[:len(idx.stack)-1]
}

func (idx *TTLIdx) PopAfter(t time.Time) []string {
	results := []string{}
	i := 0
	for _, e := range idx.stack {
		if e.until.Before(t) {
			results = append(results, e.key)
		}
		i++
	}
	if i > 0 {
		idx.stack = idx.stack[i-1:]
	}
	return results
}

type event struct {
	Record engine.Record `json:"record"`
	Action action        `json:"action"`
}

func (e *event) Bytes() []byte {
	return []byte(fmt.Sprintf("%d::%s", e.Action, e.Record))
}

func New(path string, log logger.Logger) (*Narwal, error) {
	wal, err := OpenWAL(path, DefaultMaxRecordSize, log)
	if err != nil {
		return nil, errors.Wrap(err, "open WAL")
	}
	snapshot, err := wal.Read()
	if err != nil {
		return nil, err
	}

	storage := &Narwal{
		log:    log,
		wal:    wal,
		lock:   &sync.RWMutex{},
		data:   snapshot,
		ttlIdx: &TTLIdx{stack: []ttlRecord{}},
	}

	go func() {
		t := time.NewTicker(time.Second * 3)
		for {
			select {
			case <-t.C:
				keys := storage.ttlIdx.PopAfter(time.Now())
				log.Infof("keys: %s", keys)
				for _, key := range keys {
					log.Infof("Removed by TTL: %s", key)
					storage.Delete(key)
				}
			}
		}
	}()

	return storage, nil
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
	s.log.Infof("pattern %s", regPattern)
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
	if record.ExpirationTime != nil {
		s.ttlIdx.Push(ttlRecord{key: record.Key, until: *record.ExpirationTime})
	}
	s.data[record.Key] = record
	s.log.Infof("data %v", s.data)
	s.log.Infof("ttl %v", s.ttlIdx)
}

func (s *Narwal) delete(key string) {
	if err := s.wal.Write(event{Record: engine.Record{Key: key}, Action: actionDelete}); err != nil {
		s.log.Error(err)
	}
	s.ttlIdx.Delete(key)
	delete(s.data, key)
}
