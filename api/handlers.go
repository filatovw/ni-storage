package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/render"

	"github.com/go-chi/chi"

	"github.com/filatovw/ni-storage/config"
	"github.com/filatovw/ni-storage/engine"
	"github.com/filatovw/ni-storage/logger"
)

func HealthHandler(log logger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.PlainText(w, r, http.StatusText(http.StatusOK))
	})
}

type Server struct {
	storage engine.Storage
	config  config.Config
	log     logger.Logger
}

func (s *Server) write(w http.ResponseWriter, body []byte, status int) {
	w.WriteHeader(status)
	fmt.Fprint(w, body)
}

// GetHandler get a value (GET /keys/{id})
func (s *Server) GetHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	item, ok := s.storage.Get(id)
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	render.JSON(w, r, item.Value)
}

// GetAllHandler get all values (GET /keys)
// support wildcard keys when getting all values (GET /keys?filter=wo$d)
// (the $ symbol should expand to match any number of characters, e.g: wod, word, world etc.)
func (s *Server) GetAllHandler(w http.ResponseWriter, r *http.Request) {
	var (
		records []engine.Record
		err     error
	)

	pattern := r.URL.Query().Get("filter")
	if pattern != "" {
		records, err = s.storage.Filter(pattern)
	} else {
		records, err = s.storage.GetAll()
	}
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	items := make([]string, len(records))
	for i, v := range records {
		items[i] = v.Value
	}

	render.JSON(w, r, items)
}

// SetHandler set a value (PUT /keys), set an expiry time when adding a value (PUT /keys?expire_in=60)
func (s *Server) SetHandler(w http.ResponseWriter, r *http.Request) {
	expireInParam := r.URL.Query().Get("expire_in")
	item := engine.Record{}
	if expireInParam != "" {
		expireIn, err := strconv.Atoi(expireInParam)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		ts := time.Now().Add(time.Duration(expireIn) * time.Second)
		item.ExpirationTime = &ts
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(
			w,
			http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	item.Key = id
	item.Value = string(body)
	s.storage.Set(item)
	w.WriteHeader(http.StatusCreated)

	render.JSON(w, r, http.StatusText(http.StatusOK))
}

// CheckHandler check if a value exists (HEAD /keys/{id})
func (s *Server) CheckHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if ok := s.storage.Exists(id); !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	render.JSON(w, r, true)
}

// DeleteHandler delete a value (DELETE /keys/{id})
func (s *Server) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	s.storage.Delete(id)
	w.WriteHeader(http.StatusAccepted)
	render.JSON(w, r, http.StatusText(http.StatusOK))
}

// DeleteAllHandler delete all values (DELETE /keys)
func (s *Server) DeleteAllHandler(w http.ResponseWriter, r *http.Request) {
	s.storage.DeleteAll()
	w.WriteHeader(http.StatusAccepted)
	render.JSON(w, r, http.StatusText(http.StatusOK))
}
