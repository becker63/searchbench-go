package usage

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
)

// HashRegistry records deterministic request/response hashes for live attribution (#86).
type HashRegistry struct {
	mu        sync.Mutex
	requests  []string
	responses []string
}

// RecordRequest hashes and stores one model request payload.
func (r *HashRegistry) RecordRequest(payload []byte) {
	if r == nil || len(payload) == 0 {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requests = append(r.requests, hashBytes(payload))
}

// RecordResponse hashes and stores one model response payload.
func (r *HashRegistry) RecordResponse(payload []byte) {
	if r == nil || len(payload) == 0 {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.responses = append(r.responses, hashBytes(payload))
}

// Snapshot returns copies of recorded hashes.
func (r *HashRegistry) Snapshot() (requests, responses []string) {
	if r == nil {
		return nil, nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	req := append([]string(nil), r.requests...)
	resp := append([]string(nil), r.responses...)
	return req, resp
}

func hashBytes(payload []byte) string {
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}
