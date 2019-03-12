package ni

import (
	"sync"
	"time"

	"github.com/filatovw/ni-storage/engine"
)

type Ni struct {
	// data storage
	data map[string]engine.Record
	lock *sync.RWMutex
}

func (s Ni) Get(key string) (*engine.Record, error) {
	return nil, nil
}

func (s Ni) GetAll() ([]*engine.Record, error) {
	return nil, nil
}

func (s *Ni) Set(record engine.Record) error {
	return nil
}

func (s *Ni) Delete(key string) error {
	return nil
}

func (s *Ni) DeleteAll() error {
	return nil
}

func (s *Ni) SetExpirationTime(key string, ttl time.Duration) error {
	return nil
}

func (s Ni) Filter(pattern string) ([]*engine.Record, error) {
	return nil, nil
}
