package engine

import "time"

// Null is the empty record
var Null = Record{}

// Record entity in a Storage
type Record struct {
	// ExpirationTime can be empty for records with endless existance
	ExpirationTime *time.Time `json:"expiration_time,omitempty"`
	Value          string     `json:"value,omitempty"`
	Key            string     `json:"key"`
}

// Storage simple KV-storage
type Storage interface {
	// Exists check if key exists in a storage
	Exists(string) bool
	// Get find record by key
	Get(string) (Record, bool)
	// GetAll get all records from storage
	GetAll() ([]Record, error)
	// Filter get all records passed filtering by passed pattern where "$"" means "any number of symbols"
	Filter(string) ([]Record, error)
	// Set save record in a storage
	Set(Record)
	// Delete remove record with defined key
	Delete(string)
	// DeleteALl remove all records
	DeleteAll()
}
