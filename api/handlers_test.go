package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/filatovw/ni-storage/engine"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

// MockStorage is for testing needs only
type MockStorage struct {
	data map[string]engine.Record
}

func (s MockStorage) Exists(key string) bool {
	_, ok := s.data[key]
	return ok
}

func (s MockStorage) Get(key string) (engine.Record, bool) {
	v, ok := s.data[key]
	return v, ok
}

func (s MockStorage) GetAll() map[string]engine.Record {
	return s.data
}

func (s MockStorage) Filter(pattern string) (map[string]engine.Record, error) {
	return s.data, nil
}

func (s MockStorage) Set(record engine.Record) {
	s.data[record.Key] = record
}

func (s MockStorage) Delete(key string) {
	delete(s.data, key)
}

func (s MockStorage) DeleteAll() {
	for k := range s.data {
		s.Delete(k)
	}
}

func setupServer(t *testing.T) Server {
	t.Helper()
	log, err := zap.NewProduction()
	if err != nil {
		t.Errorf("error on logger init: %s", err)
	}
	storage := MockStorage{data: make(map[string]engine.Record)}
	server := Server{storage: storage, log: log.Sugar()}
	return server
}

func TestSetHandler(t *testing.T) {
	server := setupServer(t)

	rr := httptest.NewRecorder()
	body := strings.NewReader("value1")
	req, err := http.NewRequest("PUT", "/keys/key1", body)
	if err != nil {
		t.Fatal(err)
	}
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "key1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler := http.HandlerFunc(server.SetHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := "\"OK\"\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %#v want %#v",
			rr.Body.String(), expected)
	}
}

func TestSetMultipleHandler(t *testing.T) {
	server := setupServer(t)

	td1 := time.Duration(59)
	body, err := json.Marshal(
		map[string]record{
			"key1": record{Value: "value1", ExpireIn: &td1},
			"key2": record{Value: "value2"},
		},
	)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	rb := bytes.NewReader(body)
	rr := httptest.NewRecorder()
	req, err := http.NewRequest("PUT", "/keys", rb)
	if err != nil {
		t.Fatal(err)
	}

	handler := http.HandlerFunc(server.SetMultipleHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := "\"OK\"\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %#v want %#v",
			rr.Body.String(), expected)
	}
}

func TestGetHandlerOK(t *testing.T) {
	server := setupServer(t)
	record := engine.Record{Key: "key1", Value: "value1"}
	server.storage.Set(record)

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/keys/key1", nil)
	if err != nil {
		t.Fatal(err)
	}

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "key1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler := http.HandlerFunc(server.GetHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := "\"value1\"\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %#v want %#v",
			rr.Body.String(), expected)
	}
}

func TestGetHandlerNotFound(t *testing.T) {
	server := setupServer(t)

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/keys/key1", nil)
	if err != nil {
		t.Fatal(err)
	}

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "key1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler := http.HandlerFunc(server.GetHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := "\"Not Found\"\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %#v want %#v",
			rr.Body.String(), expected)
	}
}

func TestCheckHandlerOK(t *testing.T) {
	server := setupServer(t)
	record := engine.Record{Key: "key1", Value: "value1"}
	server.storage.Set(record)

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/keys/key1", nil)
	if err != nil {
		t.Fatal(err)
	}

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "key1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler := http.HandlerFunc(server.CheckHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := "true\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %#v want %#v",
			rr.Body.String(), expected)
	}
}

func TestCheckHandlerNotFound(t *testing.T) {
	server := setupServer(t)

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/keys/key1", nil)
	if err != nil {
		t.Fatal(err)
	}

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "key1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler := http.HandlerFunc(server.CheckHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := "\"Not Found\"\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %#v want %#v",
			rr.Body.String(), expected)
	}
}

func TestDeleteHandler(t *testing.T) {
	server := setupServer(t)
	record := engine.Record{Key: "key1", Value: "value1"}
	server.storage.Set(record)

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("DELETE", "/keys/key1", nil)
	if err != nil {
		t.Fatal(err)
	}

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "key1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler := http.HandlerFunc(server.DeleteHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusAccepted {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusAccepted)
	}

	expected := "\"Accepted\"\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %#v want %#v",
			rr.Body.String(), expected)
	}
}

func TestGetAllHandler(t *testing.T) {
	server := setupServer(t)

	record1 := engine.Record{Key: "key1", Value: "value1"}
	record2 := engine.Record{Key: "key2", Value: "value2"}

	server.storage.Set(record1)
	server.storage.Set(record2)

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/keys", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler := http.HandlerFunc(server.GetAllHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusAccepted)
	}

	var v []string
	if err := json.Unmarshal(rr.Body.Bytes(), &v); err != nil {
		t.Errorf("error on unmarshalling: %s", err)
	}
	expected := []string{
		record1.Key,
		record2.Key,
	}

	sort.Slice(v, func(i, j int) bool { return v[i] < v[j] })

	if !reflect.DeepEqual(v, expected) {
		t.Errorf("handler returned unexpected body: got %#v want %#v",
			v, expected)
	}
}

func TestDeleteAllHandler(t *testing.T) {
	server := setupServer(t)
	record := engine.Record{Key: "key1", Value: "value1"}
	server.storage.Set(record)

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("DELETE", "/keys", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler := http.HandlerFunc(server.DeleteAllHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusAccepted {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusAccepted)
	}

	expected := "\"Accepted\"\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %#v want %#v",
			rr.Body.String(), expected)
	}
	if len(server.storage.GetAll()) != 0 {
		t.Errorf("expected empty storage, actual size is: %d", len(server.storage.GetAll()))
	}
}
