package narwal

import (
	"context"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/filatovw/ni-storage/engine"
	"github.com/filatovw/ni-storage/engine/narwal/ttl"
	"github.com/filatovw/ni-storage/logger"
	"github.com/pkg/errors"
)

var (
	defaultTTLCheckPeriod = 2 * time.Second
)

// Narwal engine stores data on a disk and keeps copy of data in memory.
type Narwal struct {
	log  logger.Logger
	lock *sync.RWMutex
	data map[string]engine.Record
	wal  *WAL
	ttl  *ttl.Index
}

// event holds state container and performed action
type event struct {
	Record engine.Record `json:"record"`
	Action action        `json:"action"`
}

// New creates engine object
func New(ctx context.Context, path string, log logger.Logger) (*Narwal, error) {
	wal, err := OpenWAL(log, path, defaultMaxRecordSize)
	if err != nil {
		return nil, errors.Wrap(err, "open WAL")
	}
	snapshot, err := wal.Read()
	if err != nil {
		return nil, err
	}

	ttlIndex := ttl.NewIndex()

	storage := &Narwal{
		log:  log,
		wal:  wal,
		lock: &sync.RWMutex{},
		data: snapshot,
		ttl:  &ttlIndex,
	}
	storage.deleteExpired(time.Now())

	go storage.closeWAL(ctx)
	go storage.checkExpired(ctx, defaultTTLCheckPeriod)
	return storage, nil
}

// closeWAL closes log-file
func (s *Narwal) closeWAL(ctx context.Context) {
	<-ctx.Done()
	if err := s.wal.Close(); err != nil {
		s.log.Errorf("failed to close WAL, possible data corruption: %s", err)
	}
}

// deleteExpired delete all keys that are expired by the time
func (s *Narwal) deleteExpired(t time.Time) {
	keys := s.ttl.PopAfter(t)
	s.log.Debugf("keys: %s", keys)
	for _, key := range keys {
		s.log.Debugf("Removed expired: %s", key)
		s.Delete(key)
	}
}

// checkExpired check if any records have expired time
func (s *Narwal) checkExpired(ctx context.Context, period time.Duration) {
	t := time.NewTicker(period)
	for {
		select {
		case <-t.C:
			s.deleteExpired(time.Now())
		case <-ctx.Done():
			t.Stop()
			return
		}
	}
}

// Exists check if key exists in a storage
func (s *Narwal) Exists(key string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	_, ok := s.data[key]
	return ok
}

// Get find record by key
func (s *Narwal) Get(key string) (engine.Record, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	record, ok := s.data[key]
	if !ok {
		return engine.Null, false
	}
	return record, true
}

// Set save record in a storage
func (s *Narwal) Set(record engine.Record) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.set(record)
}

// Delete remove record with defined key
func (s *Narwal) Delete(key string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.delete(key)
}

// Filter get all records passed filtering by pattern where "$"" means "any number of symbols"
func (s *Narwal) Filter(pattern string) (map[string]engine.Record, error) {
	regPattern := strings.ReplaceAll(pattern, "$", ".*")
	s.log.Debugf("pattern %s", regPattern)
	exp, err := regexp.Compile(regPattern)
	if err != nil {
		return nil, err
	}

	s.lock.RLock()
	defer s.lock.RUnlock()

	results := make(map[string]engine.Record)
	for _, v := range s.data {
		if exp.MatchString(v.Value) {
			results[v.Key] = v
		}
	}
	return results, nil
}

// GetAll get all records from storage
func (s *Narwal) GetAll() map[string]engine.Record {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.data
}

// DeleteAll remove all records
func (s *Narwal) DeleteAll() {
	s.lock.Lock()
	defer s.lock.Unlock()
	for k := range s.data {
		s.delete(k)
	}
}

// set save record in a storage
func (s *Narwal) set(record engine.Record) {
	//  check if record has already expired
	if record.ExpirationTime != nil && record.ExpirationTime.Before(time.Now()) {
		return
	}
	if err := s.wal.Write(event{Record: record, Action: actionSet}); err != nil {
		s.log.Error(err)
	}
	if record.ExpirationTime != nil {
		s.ttl.Push(ttl.Record{Key: record.Key, Until: *record.ExpirationTime})
	}
	s.data[record.Key] = record
}

// delete remove value from a storage by key
func (s *Narwal) delete(key string) {
	if err := s.wal.Write(event{Record: engine.Record{Key: key}, Action: actionDelete}); err != nil {
		s.log.Error(err)
	}
	s.ttl.Delete(key)
	delete(s.data, key)
}
