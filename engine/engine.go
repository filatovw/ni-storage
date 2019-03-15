package engine

import "time"

var Null = Record{}

type Record struct {
	ExpirationTime *time.Time
	Value          string
	Key            string
}

type Storage interface {
	Exists(string) bool
	Get(string) (Record, bool)
	GetAll() ([]Record, error)
	Filter(string) ([]Record, error)
	Set(Record)
	Delete(string)
	DeleteAll()
}
