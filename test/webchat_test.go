package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/channels"
	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestWebChatChannel(t *testing.T) {
	// 1. Setup Config e MessageBus
	cfg := config.WebChatConfig{
		Enabled: true,
		Host:    "127.0.0.1",
		Port:    4002, // Usa una porta diversa per il test per evitare conflitti
	}
	messageBus := bus.NewMessageBus()

	// 2. Crea e avvia il canale WebChat
	tempDir := t.TempDir()
	channel, err := channels.NewWebChatChannel(cfg, messageBus, tempDir, nil)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = channel.Start(ctx)
	assert.NoError(t, err)
	defer channel.Stop(ctx)

	// Aspetta che il server si avvii
	time.Sleep(100 * time.Millisecond)

	// 3. Connetti un client WebSocket
	url := fmt.Sprintf("ws://%s:%d/ws", cfg.Host, cfg.Port)
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)
	defer ws.Close()

	// 4. Invia un messaggio dal client
	testMessage := "Hello WebChat"
	err = ws.WriteJSON(map[string]interface{}{
		"content": testMessage,
	})
	assert.NoError(t, err)

	// 5. Verifica che il messaggio arrivi sul bus
	// Modifica: Usa ConsumeInbound invece di accedere direttamente al canale
	msg, ok := messageBus.ConsumeInbound(ctx)
	assert.True(t, ok)
	assert.Equal(t, "webchat", msg.Channel)
	assert.Equal(t, testMessage, msg.Content)
	assert.Equal(t, "web_user", msg.SenderID)

	// 6. Invia un messaggio di risposta tramite il canale
	responseMessage := "Hello from Bot"
	outMsg := bus.OutboundMessage{
		Channel: "webchat",
		Content: responseMessage,
	}
	err = channel.Send(ctx, outMsg)
	assert.NoError(t, err)

	// 7. Verifica che il client riceva la risposta
	var response map[string]interface{}
	err = ws.ReadJSON(&response)
	assert.NoError(t, err)
	assert.Equal(t, responseMessage, response["content"])
	assert.Equal(t, "message", response["type"])
}
