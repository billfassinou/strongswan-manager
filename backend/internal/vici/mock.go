// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

package vici

import (
	"context"
	"strings"
	"sync"

	"strongswan-manager/internal/domain"
)

// Mock est une implémentation en mémoire de l'Adapter pour les tests unitaires
// et le développement sans lab (les connexions chargées sont mémorisées).
type Mock struct {
	mu     sync.Mutex
	conns  map[string]map[string]any
	up     map[string]bool
	shared  map[string]string // id du secret -> valeur chargée (load-shared)
	certs   int
	caCerts int
	keys    int
	Ver     string
}

// NewMock crée un adaptateur mock.
func NewMock() *Mock {
	return &Mock{conns: map[string]map[string]any{}, up: map[string]bool{}, shared: map[string]string{}, Ver: "6.0.1"}
}

func (m *Mock) Version(context.Context) (string, error) { return m.Ver, nil }

func (m *Mock) LoadConn(_ context.Context, name string, conn map[string]any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.conns[name] = conn
	return nil
}

func (m *Mock) UnloadConn(_ context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.conns, name)
	delete(m.up, name)
	return nil
}

func (m *Mock) LoadShared(_ context.Context, id, _ string, _ []string, data string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shared[id] = data
	return nil
}

// HasShared indique si un secret partagé nommé a été chargé (utile aux tests).
func (m *Mock) HasShared(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.shared[id]
	return ok
}

func (m *Mock) LoadCert(_ context.Context, _ , flag string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if flag == "ca" {
		m.caCerts++
	} else {
		m.certs++
	}
	return nil
}

func (m *Mock) LoadKey(context.Context, string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.keys++
	return nil
}

// Counts expose les compteurs de chargement pour les tests.
func (m *Mock) Counts() (certs, caCerts, keys int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.certs, m.caCerts, m.keys
}

func (m *Mock) ListSAs(context.Context) ([]domain.SAState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []domain.SAState
	for name := range m.conns {
		st := domain.StatusDown
		if m.up[name] {
			st = domain.StatusUp
		}
		out = append(out, domain.SAState{Name: name, Status: st})
	}
	return out, nil
}

func (m *Mock) Initiate(_ context.Context, child string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.up[connOf(child)] = true
	return nil
}

func (m *Mock) Terminate(_ context.Context, child string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.up[connOf(child)] = false
	return nil
}

func (m *Mock) Rekey(context.Context, string) error { return nil }
func (m *Mock) Close() error                        { return nil }

// HasConn indique si une connexion nommée est actuellement chargée (utile aux tests).
func (m *Mock) HasConn(name string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.conns[name]
	return ok
}

// connOf retire le suffixe "-net" pour retrouver le nom de connexion.
func connOf(child string) string { return strings.TrimSuffix(child, "-net") }
