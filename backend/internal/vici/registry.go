package vici

import "sync"

// Registry associe un identifiant de passerelle à son adaptateur VICI.
// Sûr pour un usage concurrent (poller + handlers HTTP).
type Registry struct {
	mu       sync.RWMutex
	adapters map[string]Adapter
}

// NewRegistry crée un registre vide.
func NewRegistry() *Registry { return &Registry{adapters: map[string]Adapter{}} }

// Set enregistre l'adaptateur d'une passerelle.
func (r *Registry) Set(gatewayID string, a Adapter) {
	r.mu.Lock()
	r.adapters[gatewayID] = a
	r.mu.Unlock()
}

// Get renvoie l'adaptateur d'une passerelle (nil, false si absent).
func (r *Registry) Get(gatewayID string) (Adapter, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.adapters[gatewayID]
	return a, ok
}

// Each itère sur toutes les paires (gatewayID, adaptateur).
func (r *Registry) Each(fn func(gatewayID string, a Adapter)) {
	r.mu.RLock()
	snapshot := make(map[string]Adapter, len(r.adapters))
	for k, v := range r.adapters {
		snapshot[k] = v
	}
	r.mu.RUnlock()
	for k, v := range snapshot {
		fn(k, v)
	}
}
