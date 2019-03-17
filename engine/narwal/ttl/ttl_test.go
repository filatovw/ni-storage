package ttl

import (
	"reflect"
	"testing"
	"time"
)

func TestTTLIndexPopAfter(t *testing.T) {
	ts1 := Record{Until: time.Date(2019, 2, 20, 16, 0, 0, 0, time.UTC), Key: "key1"}
	ts2 := Record{Until: time.Date(2019, 2, 20, 17, 0, 0, 0, time.UTC), Key: "key2"}
	ts3 := Record{Until: time.Date(2020, 2, 20, 16, 0, 0, 0, time.UTC), Key: "key3"}
	ts4 := Record{Until: time.Date(2021, 2, 20, 16, 0, 0, 0, time.UTC), Key: "key4"}

	ts1_2 := ts1
	ts1_2.Until = ts1_2.Until.AddDate(0, 0, 1)
	testData := []struct {
		name          string
		input         []Record
		expectedPop   []string
		expectedIndex []Record
	}{
		{
			name:          "no values",
			input:         []Record{},
			expectedPop:   []string{},
			expectedIndex: []Record{},
		},
		{
			name:          "1 value",
			input:         []Record{ts1},
			expectedPop:   []string{"key1"},
			expectedIndex: []Record{},
		},
		{
			name:          "4 values, 2 expired, 2 stays in index",
			input:         []Record{ts1, ts2, ts3, ts4},
			expectedPop:   []string{"key1", "key2"},
			expectedIndex: []Record{ts3, ts4},
		},
		{
			name:          "2 values, not expired",
			input:         []Record{ts3, ts4},
			expectedPop:   []string{},
			expectedIndex: []Record{ts3, ts4},
		},
		{
			name:          "4 values, random pushing",
			input:         []Record{ts4, ts2, ts3, ts1},
			expectedPop:   []string{"key1", "key2"},
			expectedIndex: []Record{ts3, ts4},
		},
		{
			name:          "value added twice, should override",
			input:         []Record{ts1, ts1_2, ts4},
			expectedPop:   []string{"key1"},
			expectedIndex: []Record{ts4},
		},
	}
	for _, td := range testData {
		t.Run(td.name, func(t *testing.T) {
			idx := NewIndex()
			for _, v := range td.input {
				idx.Push(v)
			}
			found := idx.PopAfter(time.Date(2019, 2, 22, 0, 0, 0, 0, time.UTC))
			if !reflect.DeepEqual(found, td.expectedPop) {
				t.Errorf("expected found: %#v, got: %#v", td.expectedPop, found)
			}

			if !reflect.DeepEqual(idx.stack, td.expectedIndex) {
				t.Errorf("expected remains in index: %#v, got: %#v", td.expectedIndex, idx.stack)
			}
		})
	}
}

func TestTTLIndexDelete(t *testing.T) {
	ts1 := Record{Until: time.Date(2019, 2, 20, 16, 0, 0, 0, time.UTC), Key: "key1"}
	ts2 := Record{Until: time.Date(2019, 2, 20, 17, 0, 0, 0, time.UTC), Key: "key2"}
	ts3 := Record{Until: time.Date(2020, 2, 20, 16, 0, 0, 0, time.UTC), Key: "key3"}
	testData := []struct {
		name          string
		input         []Record
		keysToDelete  []string
		expectedIndex []Record
	}{
		{
			name:          "empty index",
			input:         []Record{},
			keysToDelete:  []string{"key1", "key3"},
			expectedIndex: []Record{},
		},
		{
			name:          "1 element, delete first",
			input:         []Record{ts1},
			keysToDelete:  []string{"key1"},
			expectedIndex: []Record{},
		},
		{
			name:          "3 elements, delete first",
			input:         []Record{ts1, ts2, ts3},
			keysToDelete:  []string{"key1"},
			expectedIndex: []Record{ts2, ts3},
		},
		{
			name:          "3 elements, delete middle",
			input:         []Record{ts1, ts2, ts3},
			keysToDelete:  []string{"key2"},
			expectedIndex: []Record{ts1, ts3},
		},
		{
			name:          "3 elements, delete last",
			input:         []Record{ts1, ts2, ts3},
			keysToDelete:  []string{"key3"},
			expectedIndex: []Record{ts1, ts2},
		},
		{
			name:          "3 elements, delete 2 last",
			input:         []Record{ts1, ts2, ts3},
			keysToDelete:  []string{"key2", "key3"},
			expectedIndex: []Record{ts1},
		},
	}
	for _, td := range testData {
		t.Run(td.name, func(t *testing.T) {
			idx := NewIndex()
			for _, v := range td.input {
				idx.Push(v)
			}
			for _, k := range td.keysToDelete {
				idx.Delete(k)
			}

			if !reflect.DeepEqual(idx.stack, td.expectedIndex) {
				t.Errorf("expected remains in index: %#v, got: %#v", td.expectedIndex, idx.stack)
			}
		})
	}
}
