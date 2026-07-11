// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

package vici

import (
	"context"
	"encoding/pem"
	"fmt"
	"strings"

	govici "github.com/strongswan/govici/vici"

	"strongswan-manager/internal/domain"
)

// pemToDER convertit un bloc PEM en DER (charon attend le DER via VICI). En l'absence
// d'en-tête PEM, la valeur d'origine est renvoyée telle quelle.
func pemToDER(data string) []byte {
	if block, _ := pem.Decode([]byte(data)); block != nil {
		return block.Bytes
	}
	return []byte(data)
}

// Govici est l'implémentation réelle de l'Adapter au-dessus du client officiel govici.
// Chaque appel ouvre une session courte vers le socket VICI de la passerelle, ce qui
// évite de gérer des connexions persistantes fragiles dans cette tranche verticale.
type Govici struct {
	network string
	addr    string
	socket  string
}

// New construit un adaptateur à partir d'un endpoint :
//   "unix:/var/run/charon.vici" (défaut) ou "tcp:host:port".
func New(endpoint string) (*Govici, error) {
	g := &Govici{}
	switch {
	case strings.HasPrefix(endpoint, "unix:"):
		g.socket = strings.TrimPrefix(endpoint, "unix:")
	case strings.HasPrefix(endpoint, "tcp:"):
		g.network, g.addr = "tcp", strings.TrimPrefix(endpoint, "tcp:")
	default:
		g.socket = endpoint
	}
	return g, nil
}

func (g *Govici) session() (*govici.Session, error) {
	if g.socket != "" {
		return govici.NewSession(govici.WithSocketPath(g.socket))
	}
	return govici.NewSession(govici.WithAddr(g.network, g.addr))
}

func (g *Govici) Version(ctx context.Context) (string, error) {
	s, err := g.session()
	if err != nil {
		return "", err
	}
	defer s.Close()
	resp, err := s.CommandRequest("version", govici.NewMessage())
	if err != nil {
		return "", err
	}
	return asString(resp.Get("version")), nil
}

func (g *Govici) LoadConn(ctx context.Context, name string, conn map[string]any) error {
	s, err := g.session()
	if err != nil {
		return err
	}
	defer s.Close()
	inner, err := toMessage(conn)
	if err != nil {
		return err
	}
	msg := govici.NewMessage()
	if err := msg.Set(name, inner); err != nil {
		return err
	}
	resp, err := s.CommandRequest("load-conn", msg)
	if err != nil {
		return err
	}
	return checkSuccess(resp)
}

func (g *Govici) UnloadConn(ctx context.Context, name string) error {
	s, err := g.session()
	if err != nil {
		return err
	}
	defer s.Close()
	msg := govici.NewMessage()
	_ = msg.Set("name", name)
	resp, err := s.CommandRequest("unload-conn", msg)
	if err != nil {
		return err
	}
	return checkSuccess(resp)
}

func (g *Govici) LoadShared(ctx context.Context, id, secretType string, ikeIDs []string, data string) error {
	s, err := g.session()
	if err != nil {
		return err
	}
	defer s.Close()
	msg := govici.NewMessage()
	_ = msg.Set("id", id)
	_ = msg.Set("type", secretType)
	_ = msg.Set("data", data)
	if len(ikeIDs) > 0 {
		_ = msg.Set("owners", ikeIDs)
	}
	resp, err := s.CommandRequest("load-shared", msg)
	if err != nil {
		return err
	}
	return checkSuccess(resp)
}

func (g *Govici) LoadCert(ctx context.Context, pem, flag string) error {
	s, err := g.session()
	if err != nil {
		return err
	}
	defer s.Close()
	msg := govici.NewMessage()
	_ = msg.Set("type", "x509")
	if flag == "ca" {
		_ = msg.Set("flag", "ca")
	}
	_ = msg.Set("data", string(pemToDER(pem)))
	resp, err := s.CommandRequest("load-cert", msg)
	if err != nil {
		return err
	}
	return checkSuccess(resp)
}

func (g *Govici) LoadKey(ctx context.Context, pem string) error {
	s, err := g.session()
	if err != nil {
		return err
	}
	defer s.Close()
	msg := govici.NewMessage()
	_ = msg.Set("type", "any")
	_ = msg.Set("data", string(pemToDER(pem)))
	resp, err := s.CommandRequest("load-key", msg)
	if err != nil {
		return err
	}
	return checkSuccess(resp)
}

func (g *Govici) ListSAs(ctx context.Context) ([]domain.SAState, error) {
	s, err := g.session()
	if err != nil {
		return nil, err
	}
	defer s.Close()
	messages, err := s.StreamedCommandRequest("list-sas", "list-sa", govici.NewMessage())
	if err != nil {
		return nil, err
	}
	var out []domain.SAState
	for _, msg := range messages {
		for _, name := range msg.Keys() {
			sub, ok := msg.Get(name).(*govici.Message)
			if !ok {
				continue
			}
			out = append(out, domain.SAState{Name: name, Status: mapSAState(asString(sub.Get("state")))})
		}
	}
	return out, nil
}

func (g *Govici) Initiate(ctx context.Context, child string) error  { return g.control("initiate", child) }
func (g *Govici) Terminate(ctx context.Context, child string) error { return g.control("terminate", child) }
func (g *Govici) Rekey(ctx context.Context, child string) error     { return g.control("rekey", child) }

func (g *Govici) control(cmd, child string) error {
	s, err := g.session()
	if err != nil {
		return err
	}
	defer s.Close()
	msg := govici.NewMessage()
	_ = msg.Set("child", child)
	resp, err := s.CommandRequest(cmd, msg)
	if err != nil {
		return err
	}
	return checkSuccess(resp)
}

func (g *Govici) Close() error { return nil }

// --- helpers ---

func toMessage(m map[string]any) (*govici.Message, error) {
	msg := govici.NewMessage()
	for k, v := range m {
		if sub, ok := v.(map[string]any); ok {
			nested, err := toMessage(sub)
			if err != nil {
				return nil, err
			}
			if err := msg.Set(k, nested); err != nil {
				return nil, err
			}
			continue
		}
		if err := msg.Set(k, v); err != nil {
			return nil, err
		}
	}
	return msg, nil
}

func checkSuccess(resp *govici.Message) error {
	if resp == nil {
		return nil
	}
	if asString(resp.Get("success")) == "no" {
		return fmt.Errorf("vici: %s", asString(resp.Get("errmsg")))
	}
	return nil
}

func asString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func mapSAState(state string) string {
	switch strings.ToUpper(state) {
	case "ESTABLISHED":
		return domain.StatusUp
	case "CONNECTING", "REKEYING", "REKEYED", "CREATED":
		return domain.StatusNegotiating
	case "":
		return domain.StatusUnknown
	default:
		return domain.StatusDown
	}
}
