// Package ws diffuse les événements temps réel (état des tunnels) aux clients WebSocket.
package ws

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
)

// Hub gère l'ensemble des connexions WebSocket et la diffusion (fan-out).
type Hub struct {
	mu      sync.RWMutex
	clients map[*client]struct{}
}

type client struct {
	conn *websocket.Conn
	send chan []byte
}

// NewHub crée un hub vide.
func NewHub() *Hub { return &Hub{clients: map[*client]struct{}{}} }

// Broadcast envoie un message (JSON) à tous les clients connectés (non bloquant).
func (h *Hub) Broadcast(data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients {
		select {
		case c.send <- data:
		default: // client lent : on abandonne ce message pour lui
		}
	}
}

// Handler accepte une connexion WebSocket et la sert jusqu'à sa fermeture.
func (h *Hub) Handler(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		return
	}
	c := &client{conn: conn, send: make(chan []byte, 32)}
	h.add(c)
	defer h.remove(c)

	ctx := r.Context()
	// pompe d'écriture
	go func() {
		for msg := range c.send {
			wctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := conn.Write(wctx, websocket.MessageText, msg)
			cancel()
			if err != nil {
				return
			}
		}
	}()
	// pompe de lecture : on ignore le contenu, sert à détecter la fermeture
	for {
		if _, _, err := conn.Read(ctx); err != nil {
			conn.Close(websocket.StatusNormalClosure, "")
			return
		}
	}
}

func (h *Hub) add(c *client) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
}

func (h *Hub) remove(c *client) {
	h.mu.Lock()
	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
		close(c.send)
	}
	h.mu.Unlock()
}
