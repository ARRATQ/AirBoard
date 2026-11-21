package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestNewHTTPServerSetsTimeouts(t *testing.T) {
	srv := newHTTPServer(":0", http.NewServeMux())
	timeouts := []struct {
		name string
		got  time.Duration
	}{
		{"ReadHeaderTimeout", srv.ReadHeaderTimeout},
		{"ReadTimeout", srv.ReadTimeout},
		{"WriteTimeout", srv.WriteTimeout},
		{"IdleTimeout", srv.IdleTimeout},
	}
	for _, tt := range timeouts {
		if tt.got <= 0 {
			t.Errorf("%s non défini", tt.name)
		}
	}
	if serverShutdownTimeout >= 10*time.Second {
		t.Errorf("arrêt propre (%s) au-delà du délai SIGTERM -> SIGKILL de Docker", serverShutdownTimeout)
	}
}

// Le chat passe par un WebSocket longue durée : les délais de lecture/écriture du
// serveur HTTP ne doivent pas couper la connexion une fois détournée (Hijack).
func TestWebSocketOutlivesServerTimeouts(t *testing.T) {
	upgrader := websocket.Upgrader{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		for {
			kind, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if err := conn.WriteMessage(kind, msg); err != nil {
				return
			}
		}
	})

	ts := httptest.NewUnstartedServer(handler)
	short := 100 * time.Millisecond
	ts.Config.ReadHeaderTimeout = short
	ts.Config.ReadTimeout = short
	ts.Config.WriteTimeout = short
	ts.Config.IdleTimeout = short
	ts.Start()
	defer ts.Close()

	client, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(ts.URL, "http"), nil)
	if err != nil {
		t.Fatalf("connexion WebSocket: %v", err)
	}
	defer client.Close()

	for i := 0; i < 2; i++ {
		time.Sleep(3 * short) // au-delà de tous les délais du serveur
		if err := client.WriteMessage(websocket.TextMessage, []byte("ping")); err != nil {
			t.Fatalf("écriture %d: %v", i, err)
		}
		if err := client.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
			t.Fatalf("délai client: %v", err)
		}
		if _, msg, err := client.ReadMessage(); err != nil || string(msg) != "ping" {
			t.Fatalf("écho %d: %q, %v", i, msg, err)
		}
	}
}
