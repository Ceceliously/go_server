package service

import (
	"io"
	"net/http"
	"sync"
	"encoding/json"
	"log"
)

type srv struct {
	mu *sync.RWMutex
	data map[string]string
}



func (s *srv) Register(w http.ResponseWriter,r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	req := struct{
		Name string `json:"name"`
		Password string `json:"password"`
	}{}

	defer r.Body.Close()
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Read error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("Raw body: %s", string(raw))

	if err := json.Unmarshal(raw, &req); err != nil {
		log.Printf("Unmarshal error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	if len(req.Password) == 0 || len(req.Name) == 0{
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.data[req.Name]; exists {
		w.WriteHeader(http.StatusConflict)
		return
	}

	s.data[req.Name] = req.Password

	w.WriteHeader(http.StatusOK)
}

func (s *srv) Read(w http.ResponseWriter,r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	
	s.mu.Lock()
	data := s.data
	s.mu.Unlock()

	raw, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Write(raw)
}

func New() *srv {
	return &srv{
		mu: &sync.RWMutex{},
		data: make(map[string]string),
	}
}