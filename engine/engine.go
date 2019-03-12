package engine

import "time"

type state int

const (
	Presented state = 0
	Deleted   state = 1
)

type Record struct {
	ExpirationTime *time.Time
	State          state
	Value          string
	Key            string
}

type Storage interface {
	Get(string) (*Record, error)
	GetAll() ([]*Record, error)
	Set(Record) error
	Delete(string) error
	DeleteAll() error
	SetExpirationTime(string, time.Duration) error
	Filter(string) ([]*Record, error)
}
