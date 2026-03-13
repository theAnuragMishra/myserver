package server

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/go-chi/chi/v5"
)

// ipStore is a thread-safe in-memory store mapping IP -> visit count.
type ipStore struct {
	mu     sync.RWMutex
	counts map[string]*atomic.Int64
}

var globalIPStore = &ipStore{counts: make(map[string]*atomic.Int64)}

// record increments (or initialises) the counter for ip and returns the new value.
func (s *ipStore) record(ip string) int64 {
	// Fast path: counter already exists.
	s.mu.RLock()
	counter, ok := s.counts[ip]
	s.mu.RUnlock()
	if ok {
		return counter.Add(1)
	}

	// Slow path: create counter.
	s.mu.Lock()
	// Double-check after acquiring write lock.
	if counter, ok = s.counts[ip]; !ok {
		counter = &atomic.Int64{}
		s.counts[ip] = counter
	}
	s.mu.Unlock()
	return counter.Add(1)
}

// snapshot returns a copy of the current counts map.
func (s *ipStore) snapshot() map[string]int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]int64, len(s.counts))
	for ip, c := range s.counts {
		out[ip] = c.Load()
	}
	return out
}

// realIP extracts the best-effort real client IP from the request, honouring
// X-Forwarded-For and X-Real-IP headers set by reverse proxies (Caddy).
func realIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For may be a comma-separated list; take the first entry.
		if idx := strings.Index(xff, ","); idx != -1 {
			xff = xff[:idx]
		}
		if ip := strings.TrimSpace(xff); ip != "" {
			return ip
		}
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// IPTrackerModule registers the /tracker routes under the router.
var IPTrackerModule RouterFunc = func(r chi.Router) {
	r.Route("/tracker", func(r chi.Router) {
		// POST /tracker/visit — record a visit from the caller's IP.
		r.Post("/visit", handleRecordVisit)
		r.Group(func(r chi.Router) {
			r.Use(adminVerificationMiddleware)
			// GET  /tracker/visits — return all recorded IPs and their counts.
			r.Get("/visits", handleGetVisits)
		})
	})
}

func handleRecordVisit(w http.ResponseWriter, r *http.Request) {
	ip := realIP(r)
	count := globalIPStore.record(ip)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"ip":     ip,
		"visits": count,
	})
}

func handleGetVisits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(globalIPStore.snapshot())
}
