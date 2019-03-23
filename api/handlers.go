package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/render"

	"github.com/go-chi/chi"

	"github.com/filatovw/ni-storage/engine"
	"github.com/filatovw/ni-storage/logger"
)

// HealthHandler is used for simple health-check/echo requests
func HealthHandler(log logger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.PlainText(w, r, http.StatusText(http.StatusOK))
	})
}

type Server struct {
	storage engine.Storage
	log     logger.Logger
}

// GetHandler get a value (GET /keys/{id})
func (s *Server) GetHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	item, ok := s.storage.Get(id)
	if !ok {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, http.StatusText(http.StatusNotFound))
		return
	}
	render.JSON(w, r, item.Value)
}

// GetAllHandler get all values (GET /keys)
// support wildcard keys when getting all values (GET /keys?filter=wo$d)
// (the $ symbol should expand to match any number of characters, e.g: wod, word, world etc.)
func (s *Server) GetAllHandler(w http.ResponseWriter, r *http.Request) {
	var (
		records map[string]engine.Record
		err     error
	)

	pattern := r.URL.Query().Get("filter")
	if pattern != "" {
		records, err = s.storage.Filter(pattern)
		if err != nil {
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, http.StatusText(http.StatusInternalServerError))
			return
		}
	} else {
		records = s.storage.GetAll()
	}

	keys := make([]string, len(records))
	i := 0
	for k := range records {
		keys[i] = k
		i++
	}

	render.JSON(w, r, keys)
}

// SetHandler set a value (PUT /keys/{id}), set an expiry time when adding a value (PUT /keys?expire_in=60)
func (s *Server) SetHandler(w http.ResponseWriter, r *http.Request) {
	expireInParam := r.URL.Query().Get("expire_in")
	item := engine.Record{}
	if expireInParam != "" {
		expireIn, err := strconv.Atoi(expireInParam)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, http.StatusText(http.StatusBadRequest))
			return
		}
		ts := time.Now().Add(time.Duration(expireIn) * time.Second)
		item.ExpirationTime = &ts
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, http.StatusText(http.StatusBadRequest))
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, http.StatusText(http.StatusBadRequest))
		return
	}

	item.Key = id
	item.Value = string(body)
	s.storage.Set(item)
	w.WriteHeader(http.StatusCreated)

	render.JSON(w, r, http.StatusText(http.StatusOK))
}

type record struct {
	Value    string         `json:"value"`
	ExpireIn *time.Duration `json:"expire_in,omitempty"`
}

// SetMultipleHandler set multiple values at once (PUT /keys)
func (s *Server) SetMultipleHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, http.StatusText(http.StatusBadRequest))
		return
	}

	var req map[string]record
	if err := json.Unmarshal(body, &req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, http.StatusText(http.StatusBadRequest))
		return
	}

	tsNow := time.Now()
	for k, v := range req {
		item := engine.Record{
			Key:   k,
			Value: v.Value,
		}

		if v.ExpireIn != nil {
			ts := tsNow.Add(*v.ExpireIn * time.Second)
			item.ExpirationTime = &ts
		}
		s.storage.Set(item)
	}
	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, http.StatusText(http.StatusOK))
}

// CheckHandler check if a value exists (HEAD /keys/{id})
func (s *Server) CheckHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if ok := s.storage.Exists(id); !ok {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, http.StatusText(http.StatusNotFound))
		return
	}
	render.JSON(w, r, true)
}

// DeleteHandler delete a value (DELETE /keys/{id})
func (s *Server) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	s.storage.Delete(id)
	render.Status(r, http.StatusAccepted)
	render.JSON(w, r, http.StatusText(http.StatusAccepted))
}

// DeleteAllHandler delete all values (DELETE /keys)
func (s *Server) DeleteAllHandler(w http.ResponseWriter, r *http.Request) {
	s.storage.DeleteAll()
	render.Status(r, http.StatusAccepted)
	render.JSON(w, r, http.StatusText(http.StatusAccepted))
}
