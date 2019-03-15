package engine

import "time"

var Null = Record{}

type Record struct {
	ExpirationTime *time.Time `json:"expiration_time,omitempty"`
	Value          string     `json:"value,omitempty"`
	Key            string     `json:"key"`
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
