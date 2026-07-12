// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"time"

	"github.com/google/uuid"

	"strongswan-manager/internal/domain"
	"strongswan-manager/internal/store"
)

// --- Fakes en mémoire implémentant les interfaces de repositories ---

type fakeUsers struct{ byIdentity map[string]*domain.User }

func (f *fakeUsers) GetByIdentity(_ context.Context, id string) (*domain.User, error) {
	if u, ok := f.byIdentity[id]; ok {
		return u, nil
	}
	return nil, store.ErrNotFound
}

func (f *fakeUsers) GetByID(_ context.Context, id string) (*domain.User, error) {
	for _, u := range f.byIdentity {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, store.ErrNotFound
}

func (f *fakeUsers) SetPassword(_ context.Context, id, hash string) error {
	for _, u := range f.byIdentity {
		if u.ID == id {
			u.PassHash = hash
			u.MustChangePassword = false
			return nil
		}
	}
	return store.ErrNotFound
}

type fakeGateways struct{ m map[string]*domain.Gateway }

func (f *fakeGateways) Get(_ context.Context, id string) (*domain.Gateway, error) {
	if g, ok := f.m[id]; ok {
		return g, nil
	}
	return nil, store.ErrNotFound
}

func (f *fakeGateways) List(_ context.Context) ([]domain.Gateway, error) {
	out := make([]domain.Gateway, 0, len(f.m))
	for _, g := range f.m {
		out = append(out, *g)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

type fakeTunnels struct{ m map[string]*domain.Tunnel }

func (f *fakeTunnels) List(_ context.Context) ([]domain.Tunnel, error) {
	out := make([]domain.Tunnel, 0, len(f.m))
	for _, t := range f.m {
		out = append(out, *t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (f *fakeTunnels) Get(_ context.Context, id string) (*domain.Tunnel, error) {
	if t, ok := f.m[id]; ok {
		cp := *t
		return &cp, nil
	}
	return nil, store.ErrNotFound
}

func (f *fakeTunnels) Create(_ context.Context, t *domain.Tunnel) error {
	for _, e := range f.m {
		if e.GatewayID == t.GatewayID && e.Name == t.Name {
			return store.ErrNotFound // simule une violation d'unicité (conflit)
		}
	}
	cp := *t
	f.m[t.ID] = &cp
	return nil
}

func (f *fakeTunnels) Update(_ context.Context, t *domain.Tunnel) error {
	if _, ok := f.m[t.ID]; !ok {
		return store.ErrNotFound
	}
	cp := *t
	f.m[t.ID] = &cp
	return nil
}

func (f *fakeTunnels) Delete(_ context.Context, id string) error {
	if _, ok := f.m[id]; !ok {
		return store.ErrNotFound
	}
	delete(f.m, id)
	return nil
}

type fakeVersions struct {
	byTunnel map[string][]domain.ConfigVersion
}

func (f *fakeVersions) Create(_ context.Context, v *domain.ConfigVersion) error {
	f.byTunnel[v.TunnelID] = append(f.byTunnel[v.TunnelID], *v)
	return nil
}

func (f *fakeVersions) ListByTunnel(_ context.Context, tunnelID string) ([]domain.ConfigVersion, error) {
	vs := append([]domain.ConfigVersion{}, f.byTunnel[tunnelID]...)
	sort.Slice(vs, func(i, j int) bool { return vs[i].N > vs[j].N })
	return vs, nil
}

func (f *fakeVersions) Get(_ context.Context, tunnelID string, n int) (*domain.ConfigVersion, error) {
	for i := range f.byTunnel[tunnelID] {
		if f.byTunnel[tunnelID][i].N == n {
			v := f.byTunnel[tunnelID][i]
			return &v, nil
		}
	}
	return nil, store.ErrNotFound
}

type fakeSecrets struct {
	byName map[string]*domain.Secret
	byID   map[string]*domain.Secret
}

func newFakeSecrets() *fakeSecrets {
	return &fakeSecrets{byName: map[string]*domain.Secret{}, byID: map[string]*domain.Secret{}}
}

func (f *fakeSecrets) Create(_ context.Context, s *domain.Secret) error {
	if _, ok := f.byName[s.Name]; ok {
		return store.ErrNotFound // simule un conflit d'unicité
	}
	cp := *s
	f.byName[s.Name] = &cp
	f.byID[s.ID] = &cp
	return nil
}

func (f *fakeSecrets) GetByName(_ context.Context, name string) (*domain.Secret, error) {
	if s, ok := f.byName[name]; ok {
		return s, nil
	}
	return nil, store.ErrNotFound
}

func (f *fakeSecrets) List(_ context.Context) ([]domain.Secret, error) {
	out := make([]domain.Secret, 0, len(f.byID))
	for _, s := range f.byID {
		out = append(out, domain.Secret{ID: s.ID, Name: s.Name, Type: s.Type, UsedBy: s.UsedBy})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (f *fakeSecrets) Delete(_ context.Context, id string) error {
	s, ok := f.byID[id]
	if !ok {
		return store.ErrNotFound
	}
	delete(f.byID, id)
	delete(f.byName, s.Name)
	return nil
}

type fakeCA struct{ ca *domain.CA }

func (f *fakeCA) Get(context.Context) (*domain.CA, error) {
	if f.ca == nil {
		return nil, store.ErrNotFound
	}
	return f.ca, nil
}

func (f *fakeCA) UpdateCRL(_ context.Context, _ string, number int64, crlPEM []byte) error {
	f.ca.CRLNumber = number
	f.ca.CRLPEM = crlPEM
	return nil
}

type fakeCerts struct {
	byName map[string]*domain.Certificate
	byID   map[string]*domain.Certificate
}

func newFakeCerts() *fakeCerts {
	return &fakeCerts{byName: map[string]*domain.Certificate{}, byID: map[string]*domain.Certificate{}}
}

func (f *fakeCerts) Create(_ context.Context, c *domain.Certificate) error {
	if _, ok := f.byName[c.Name]; ok {
		return store.ErrNotFound
	}
	cp := *c
	f.byName[c.Name] = &cp
	f.byID[c.ID] = &cp
	return nil
}

func (f *fakeCerts) GetByName(_ context.Context, name string) (*domain.Certificate, error) {
	if c, ok := f.byName[name]; ok {
		return c, nil
	}
	return nil, store.ErrNotFound
}

func (f *fakeCerts) List(_ context.Context) ([]domain.Certificate, error) {
	out := make([]domain.Certificate, 0, len(f.byID))
	for _, c := range f.byID {
		out = append(out, domain.Certificate{ID: c.ID, Name: c.Name, CN: c.CN, Kind: c.Kind, Serial: c.Serial, Status: c.Status, NotBefore: c.NotBefore, NotAfter: c.NotAfter})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (f *fakeCerts) Revoke(_ context.Context, id string) error {
	c, ok := f.byID[id]
	if !ok {
		return store.ErrNotFound
	}
	c.Status = domain.CertRevoked
	return nil
}

func (f *fakeCerts) ListRevoked(context.Context) ([]domain.RevokedCert, error) {
	var out []domain.RevokedCert
	for _, c := range f.byID {
		if c.Status == domain.CertRevoked {
			out = append(out, domain.RevokedCert{Serial: c.Serial, RevokedAt: time.Now()})
		}
	}
	return out, nil
}

type fakeConfig struct{ items map[string]*domain.ConfigItem }

func newFakeConfig() *fakeConfig { return &fakeConfig{items: map[string]*domain.ConfigItem{}} }

func (f *fakeConfig) Create(_ context.Context, c *domain.ConfigItem) error {
	for _, e := range f.items {
		if e.Kind == c.Kind && e.Name == c.Name {
			return store.ErrNotFound // conflit d'unicité (kind,name)
		}
	}
	cp := *c
	f.items[c.ID] = &cp
	return nil
}
func (f *fakeConfig) List(_ context.Context, kind string) ([]domain.ConfigItem, error) {
	var out []domain.ConfigItem
	for _, c := range f.items {
		if c.Kind == kind {
			out = append(out, *c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}
func (f *fakeConfig) Get(_ context.Context, id string) (*domain.ConfigItem, error) {
	if c, ok := f.items[id]; ok {
		return c, nil
	}
	return nil, store.ErrNotFound
}
func (f *fakeConfig) Update(_ context.Context, id, name string, data json.RawMessage) error {
	c, ok := f.items[id]
	if !ok {
		return store.ErrNotFound
	}
	c.Name = name
	c.Data = data
	return nil
}
func (f *fakeConfig) Delete(_ context.Context, id string) error {
	if _, ok := f.items[id]; !ok {
		return store.ErrNotFound
	}
	delete(f.items, id)
	return nil
}

type fakeAudit struct{ entries []domain.AuditEntry }

func (f *fakeAudit) Append(_ context.Context, actorID, action, target string) error {
	var prev string
	if n := len(f.entries); n > 0 {
		prev = f.entries[n-1].IntegrityHash
	}
	h := sha256.Sum256([]byte(prev + "|" + actorID + "|" + action + "|" + target))
	f.entries = append(f.entries, domain.AuditEntry{
		ID: uuid.NewString(), ActorID: actorID, Action: action, Target: target,
		Timestamp: time.Now(), PrevHash: prev, IntegrityHash: hex.EncodeToString(h[:]),
	})
	return nil
}

func (f *fakeAudit) List(_ context.Context, limit int) ([]domain.AuditEntry, error) {
	out := append([]domain.AuditEntry{}, f.entries...)
	// plus récentes d'abord
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}
