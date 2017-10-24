package http

import (
	"net/http"
	"sync"
)

type HTTP struct {
	Data map[string]string
	mu   sync.RWMutex
}

func (h *HTTP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if len(r.RequestURI) < 2 {
		http.Error(w, "uri too short", http.StatusBadRequest)
		return
	}

	h.mu.RLock()
	val, ok := h.Data[r.RequestURI[1:]]
	h.mu.RUnlock()
	if !ok {
		http.Error(w, "key not found", http.StatusNotFound)
		return
	}

	w.Write([]byte(val))
}

func (h *HTTP) Set(key string, value string) {
	h.mu.Lock()
	h.Data[key] = value
	h.mu.Unlock()
}

func (h *HTTP) Delete(key string) {
	h.mu.Lock()
	delete(h.Data, key)
	h.mu.Unlock()
}
