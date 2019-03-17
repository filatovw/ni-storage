package narwal

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/filatovw/ni-storage/engine"
	"go.uber.org/zap"
)

func SetupEngineHelper(t *testing.T) (*Narwal, string) {
	t.Helper()
	tmpdir, err := ioutil.TempDir("", "engine_test")
	if err != nil {
		log.Fatal(err)
	}
	log, err := zap.NewProduction()
	if err != nil {
		t.Errorf("error on logger init: %s", err)
	}

	s, err := New(context.TODO(), tmpdir, log.Sugar())
	if err != nil {
		t.Errorf("create engine: %s", err)
	}
	return s, tmpdir
}

func TestEngineSet(t *testing.T) {
	s, tmpdir := SetupEngineHelper(t)
	defer os.RemoveAll(tmpdir)

	record := engine.Record{Key: "key1", Value: "value1"}
	s.Set(record)
	records := s.GetAll()
	v, ok := records[record.Key]
	if !ok {
		t.Errorf("record not found: %s", record.Key)
	}

	if !reflect.DeepEqual(v, record) {
		t.Errorf("expected: %v, got %v", record, v)
	}
}

func TestEngineSetExpiredTTL(t *testing.T) {
	s, tmpdir := SetupEngineHelper(t)
	defer os.RemoveAll(tmpdir)

	ts := time.Now()
	ts = ts.Add(-time.Second * 5)
	record := engine.Record{Key: "key1", Value: "value1", ExpirationTime: &ts}
	s.Set(record)
	expected := make(map[string]engine.Record)
	if !reflect.DeepEqual(s.GetAll(), expected) {
		t.Errorf("expected: %v, got: %v", expected, s.GetAll())
	}
}

func TestEngineSetExpiration(t *testing.T) {
	s, tmpdir := SetupEngineHelper(t)
	defer os.RemoveAll(tmpdir)

	ts := time.Now()
	record := engine.Record{Key: "key1", Value: "value1", ExpirationTime: &ts}
	s.Set(record)
	s.deleteExpired(ts.Add(-time.Second))
	expected := make(map[string]engine.Record)
	if !reflect.DeepEqual(s.GetAll(), expected) {
		t.Errorf("expected: %v, got: %v", expected, s.GetAll())
	}
}

func TestEngineGet(t *testing.T) {
	s, tmpdir := SetupEngineHelper(t)
	defer os.RemoveAll(tmpdir)

	record := engine.Record{Key: "key1", Value: "value1"}
	s.Set(record)
	found, ok := s.Get(record.Key)
	if !ok {
		t.Errorf("record with key %s not found", record.Key)
	}
	if !reflect.DeepEqual(found, record) {
		t.Errorf("expected: %v, got %v", record, found)
	}
}

func TestEngineDelete(t *testing.T) {
	s, tmpdir := SetupEngineHelper(t)
	defer os.RemoveAll(tmpdir)

	record := engine.Record{Key: "key1", Value: "value1"}
	s.Set(record)
	s.Delete(record.Key)
	expected := make(map[string]engine.Record)
	if !reflect.DeepEqual(s.GetAll(), expected) {
		t.Errorf("expected: %v, got: %v", expected, s.GetAll())
	}
}

func TestEngineExists(t *testing.T) {
	s, tmpdir := SetupEngineHelper(t)
	defer os.RemoveAll(tmpdir)

	record := engine.Record{Key: "key1", Value: "value1"}
	s.Set(record)
	ok := s.Exists(record.Key)
	if !ok {
		t.Errorf("record with key %s not found", record.Key)
	}
}

func TestEngineDeleteAll(t *testing.T) {
	s, tmpdir := SetupEngineHelper(t)
	defer os.RemoveAll(tmpdir)

	records := []engine.Record{
		engine.Record{Key: "key1", Value: "value1"},
		engine.Record{Key: "key2", Value: "value2"},
		engine.Record{Key: "key3", Value: "value3"},
	}
	for _, r := range records {
		s.Set(r)
	}
	if len(s.GetAll()) != len(records) {
		t.Errorf("unexpected number of records in a storage. Expected: %d, got: %d", len(records), len(s.GetAll()))
	}
	s.DeleteAll()
	expected := make(map[string]engine.Record)
	if !reflect.DeepEqual(s.GetAll(), expected) {
		t.Errorf("expected: %v, got: %v", expected, s.GetAll())
	}
}

func TestEngineGetAll(t *testing.T) {
	s, tmpdir := SetupEngineHelper(t)
	defer os.RemoveAll(tmpdir)

	records := map[string]engine.Record{
		"key1": engine.Record{Key: "key1", Value: "value1"},
		"key2": engine.Record{Key: "key2", Value: "value2"},
		"key3": engine.Record{Key: "key3", Value: "value3"},
	}
	for _, r := range records {
		s.Set(r)
	}
	if !reflect.DeepEqual(s.GetAll(), records) {
		t.Errorf("expected: %v, got: %v", records, s.GetAll())
	}
}
