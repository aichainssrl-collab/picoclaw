package channels

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/logger"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Permetti tutte le origini per semplicità in sviluppo
	},
}

type WebChatChannel struct {
	*BaseChannel
	config    config.WebChatConfig
	workspace string
	server    *http.Server
	clients   map[*websocket.Conn]bool
	mu        sync.Mutex
}

func NewWebChatChannel(cfg config.WebChatConfig, messageBus *bus.MessageBus, workspace string) (*WebChatChannel, error) {
	base := NewBaseChannel("webchat", cfg, messageBus, cfg.AllowFrom)
	return &WebChatChannel{
		BaseChannel: base,
		config:      cfg,
		workspace:   workspace,
		clients:     make(map[*websocket.Conn]bool),
	}, nil
}

func (c *WebChatChannel) Start(ctx context.Context) error {
	c.running = true
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", c.handleWebSocket)
	
	// Aggiungi una semplice pagina HTML per il test se richiesto, o servi file statici
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "webchat_client.html")
	})

	// Serve static assets
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	// Endpoint per le news AI
	mux.HandleFunc("/api/news", func(w http.ResponseWriter, r *http.Request) {
		newsPath := filepath.Join(c.workspace, "news", "news.json")
		data, err := os.ReadFile(newsPath)
		if err != nil {
			logger.ErrorCF("channels", "Failed to read news file", map[string]interface{}{
				"path":  newsPath,
				"error": err.Error(),
			})
			http.Error(w, "Failed to load news", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	})

	// Endpoint per i dati Crypto
	mux.HandleFunc("/api/crypto", func(w http.ResponseWriter, r *http.Request) {
		cryptoPath := filepath.Join(c.workspace, "news", "crypto.json")
		data, err := os.ReadFile(cryptoPath)
		if err != nil {
			logger.ErrorCF("channels", "Failed to read crypto file", map[string]interface{}{
				"path":  cryptoPath,
				"error": err.Error(),
			})
			http.Error(w, "Failed to load crypto data", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	})

	// Endpoint per i dati Analytics Social
	mux.HandleFunc("/api/analytics", func(w http.ResponseWriter, r *http.Request) {
		socialPath := filepath.Join(c.workspace, "social.json")
		data, err := os.ReadFile(socialPath)
		if err != nil {
			logger.ErrorCF("channels", "Failed to read social file", map[string]interface{}{
				"path":  socialPath,
				"error": err.Error(),
			})
			// Return empty array if file doesn't exist yet
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("[]"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	})

	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
	c.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	logger.InfoCF("channels", "WebChat channel starting on %s", map[string]interface{}{
		"address": addr,
	})

	go func() {
		if err := c.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.ErrorCF("channels", "WebChat server error", map[string]interface{}{
				"error": err.Error(),
			})
			c.running = false
		}
	}()

	return nil
}

func (c *WebChatChannel) Stop(ctx context.Context) error {
	c.running = false
	if c.server != nil {
		return c.server.Shutdown(ctx)
	}
	return nil
}

func (c *WebChatChannel) Send(ctx context.Context, msg bus.OutboundMessage) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Invia il messaggio a tutti i client connessi
	// In una implementazione reale, dovremmo filtrare per ChatID/SenderID
	for client := range c.clients {
		err := client.WriteJSON(map[string]interface{}{
			"type":    "message",
			"content": msg.Content,
			"sender":  "bot",
		})
		if err != nil {
			logger.ErrorCF("channels", "Error sending message to webchat client", map[string]interface{}{
				"error": err.Error(),
			})
			client.Close()
			delete(c.clients, client)
		}
	}
	return nil
}

func (c *WebChatChannel) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.ErrorCF("channels", "Failed to upgrade websocket", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	c.mu.Lock()
	c.clients[conn] = true
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.clients, conn)
		c.mu.Unlock()
		conn.Close()
	}()

	for {
		var msg map[string]interface{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			break
		}

		content, ok := msg["content"].(string)
		if !ok {
			continue
		}

		// Simula un sender ID e chat ID per la webchat
		senderID := "web_user"
		chatID := "web_chat"

		c.HandleMessage(senderID, chatID, content, nil, nil)
	}
}
