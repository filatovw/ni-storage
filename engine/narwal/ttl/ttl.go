package ttl

// TTL index serves for storing of <time:key> pair.
// The underlying struct is a stack. Older values lie on top which makes it easier to check just a few of them regularely.
// Downside of this approach is the complexity of adding of a new value or deletion by key
// but in context of KV database with in-memory cache Reads happens way more often than Writes.

import (
	"sort"
	"time"
)

// Index stores data in a stack-like structure and provides Push, PopAfter and Delete operations.
type Index struct {
	stack []Record
}

// NewIndex creates TTL index
func NewIndex() Index {
	return Index{
		stack: []Record{},
	}
}

// Record keeps time of expiration and Key of a record
type Record struct {
	Until time.Time
	Key   string
}

// Push record on stack
func (idx *Index) Push(r Record) {
	idx.stack = append(idx.stack, r)
	sort.Slice(idx.stack, func(i, j int) bool { return idx.stack[i].Until.Before(idx.stack[j].Until) })
}

// Delete record by key
func (idx *Index) Delete(k string) {
	i := -1
	for j, e := range idx.stack {
		if e.Key == k {
			i = j
			break
		}
	}
	if i >= 0 {
		idx.stack = append(idx.stack[:i], idx.stack[i+1:]...)
	}
}

// PopAfter returns all the keys with expiration time older than given time
func (idx *Index) PopAfter(t time.Time) []string {
	results := []string{}
	i := 0
	for _, e := range idx.stack {
		if e.Until.Before(t) {
			results = append(results, e.Key)
			i++
		} else {
			break
		}
	}
	if i > 0 {
		idx.stack = idx.stack[i:]
	}
	return results
}
