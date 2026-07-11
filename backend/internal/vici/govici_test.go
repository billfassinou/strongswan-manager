// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

package vici

import (
	"testing"

	govici "github.com/strongswan/govici/vici"

	"strongswan-manager/internal/domain"
)

func TestToMessageNested(t *testing.T) {
	conn := map[string]any{
		"version":     "2",
		"local_addrs": []string{"203.0.113.10"},
		"local":       map[string]any{"auth": "psk"},
		"children": map[string]any{
			"net": map[string]any{"esp_proposals": []string{"aes256gcm16"}},
		},
	}
	msg, err := toMessage(conn)
	if err != nil {
		t.Fatal(err)
	}
	if msg.Get("version") != "2" {
		t.Fatalf("version = %v", msg.Get("version"))
	}
	local, ok := msg.Get("local").(*govici.Message)
	if !ok || local.Get("auth") != "psk" {
		t.Fatalf("section imbriquée 'local' incorrecte: %v", msg.Get("local"))
	}
	children, ok := msg.Get("children").(*govici.Message)
	if !ok {
		t.Fatal("section 'children' manquante")
	}
	if _, ok := children.Get("net").(*govici.Message); !ok {
		t.Fatal("section 'children.net' manquante")
	}
}

func TestMapSAState(t *testing.T) {
	cases := map[string]string{
		"ESTABLISHED": domain.StatusUp,
		"CONNECTING":  domain.StatusNegotiating,
		"REKEYING":    domain.StatusNegotiating,
		"DELETING":    domain.StatusDown,
		"":            domain.StatusUnknown,
	}
	for in, want := range cases {
		if got := mapSAState(in); got != want {
			t.Fatalf("mapSAState(%q)=%q, attendu %q", in, got, want)
		}
	}
}

func TestCheckSuccess(t *testing.T) {
	ok := govici.NewMessage()
	_ = ok.Set("success", "yes")
	if err := checkSuccess(ok); err != nil {
		t.Fatalf("succès attendu: %v", err)
	}

	ko := govici.NewMessage()
	_ = ko.Set("success", "no")
	_ = ko.Set("errmsg", "boom")
	if err := checkSuccess(ko); err == nil {
		t.Fatal("échec attendu")
	}

	if err := checkSuccess(nil); err != nil {
		t.Fatalf("nil devrait être sans erreur: %v", err)
	}
}

func TestNewEndpointParsing(t *testing.T) {
	g, _ := New("unix:/var/run/charon.vici")
	if g.socket != "/var/run/charon.vici" || g.network != "" {
		t.Fatalf("parsing unix incorrect: %+v", g)
	}
	g, _ = New("tcp:host:4502")
	if g.network != "tcp" || g.addr != "host:4502" {
		t.Fatalf("parsing tcp incorrect: %+v", g)
	}
	g, _ = New("/chemin/direct.vici")
	if g.socket != "/chemin/direct.vici" {
		t.Fatalf("parsing défaut incorrect: %+v", g)
	}
}

func TestAsString(t *testing.T) {
	if asString("x") != "x" || asString(42) != "" || asString(nil) != "" {
		t.Fatal("asString incorrect")
	}
}
