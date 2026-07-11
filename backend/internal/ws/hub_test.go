// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

package ws

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

func TestHubBroadcast(t *testing.T) {
	hub := NewHub()
	srv := httptest.NewServer(http.HandlerFunc(hub.Handler))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	// laisse le hub enregistrer le client
	waitClients(hub, 1)

	hub.Broadcast([]byte(`{"type":"tunnel_status","status":"up"}`))

	_, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(string(data), `"status":"up"`) {
		t.Fatalf("message inattendu: %s", data)
	}
}

func TestHubRemovesClientOnClose(t *testing.T) {
	hub := NewHub()
	srv := httptest.NewServer(http.HandlerFunc(hub.Handler))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	waitClients(hub, 1)
	conn.Close(websocket.StatusNormalClosure, "")
	waitClients(hub, 0)
}

func waitClients(h *Hub, n int) {
	for i := 0; i < 100; i++ {
		h.mu.RLock()
		c := len(h.clients)
		h.mu.RUnlock()
		if c == n {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}
