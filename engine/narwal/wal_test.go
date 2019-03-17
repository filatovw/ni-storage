package narwal

import (
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/filatovw/ni-storage/engine"
)

func TestWAL(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "wal_test")
	if err != nil {
		log.Fatal(err)
	}

	defer os.RemoveAll(tmpdir) // clean up

	// create new WAL
	wal, err := OpenWAL(nil, tmpdir, 2<<10)
	if err != nil {
		t.Errorf("error on open: %s", err)
		return
	}
	record1 := engine.Record{Key: "key1", Value: "value1"}
	record2 := engine.Record{Key: "key2", Value: "value2"}
	ts := time.Date(2059, 1, 1, 1, 1, 1, 0, time.UTC)
	record3 := engine.Record{Key: "key3", Value: "value3", ExpirationTime: &ts}
	record4 := engine.Record{Key: "key1"}

	input := []event{
		{
			Record: record1,
			Action: actionSet,
		},
		{
			Record: record2,
			Action: actionSet,
		},
		{
			Record: record3,
			Action: actionSet,
		},
		{
			Record: record4,
			Action: actionDelete,
		},
	}
	expected := map[string]engine.Record{
		"key2": record2,
		"key3": record3,
	}
	for _, e := range input {
		if err := wal.Write(e); err != nil {
			t.Errorf("error on writing: %s", err)
			return
		}
	}

	if err := wal.Close(); err != nil {
		t.Errorf("error on closing: %s", err)
		return
	}

	// open existing WAL
	wal, err = OpenWAL(nil, tmpdir, 2<<10)
	if err != nil {
		t.Errorf("error on open: %s", err)
		return
	}
	snapshot, err := wal.Read()
	if err != nil {
		t.Errorf("error on reading WAL: %s", err)
	}
	if !reflect.DeepEqual(snapshot, expected) {
		t.Errorf("expected: %s, got: %s", expected, snapshot)
	}
}

func TestWALWrongPermissions(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "wal_perm_test")
	if err != nil {
		log.Fatal(err)
	}

	defer os.RemoveAll(tmpdir) // clean up
	if err := os.Chmod(tmpdir, 0444); err != nil {
		t.Errorf("permission update: %s", err)
	}

	// create new WAL
	_, err = OpenWAL(nil, tmpdir, 2<<10)
	if err == nil {
		t.Errorf("expected permission error")
		return
	}
}

func TestWALValueTooLarge(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "wal_too_large_test")
	if err != nil {
		log.Fatal(err)
	}

	defer os.RemoveAll(tmpdir) // clean up

	// create new WAL
	wal, err := OpenWAL(nil, tmpdir, 2)
	if err != nil {
		t.Errorf("error on open: %s", err)
		return
	}
	err = wal.Write(event{Action: actionSet, Record: engine.Record{Key: "some key", Value: "123"}})
	if err == nil {
		t.Errorf("expected error: too large value, got nothing")
		return
	}
}
