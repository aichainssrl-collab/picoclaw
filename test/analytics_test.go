package test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/channels"
	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestWebChatAnalytics(t *testing.T) {
	// 1. Setup Config e MessageBus
	cfg := config.WebChatConfig{
		Enabled: true,
		Host:    "127.0.0.1",
		Port:    4003, // Usa una porta diversa per il test per evitare conflitti
	}
	messageBus := bus.NewMessageBus()

	// 2. Setup directory temporanea e file social.json
	tempDir := t.TempDir()
	socialData := `[
		{
			"date": "2026-02-19",
			"facebook_followers": 262,
			"instagram_followers": 1,
			"instagram_posts": 2,
			"linkedin_followers": 7,
			"twitter_followers": 0,
			"youtube_subscribers": 0
		}
	]`
	socialPath := filepath.Join(tempDir, "social.json")
	err := os.WriteFile(socialPath, []byte(socialData), 0644)
	assert.NoError(t, err)

	// 3. Crea e avvia il canale WebChat
	channel, err := channels.NewWebChatChannel(cfg, messageBus, tempDir)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = channel.Start(ctx)
	assert.NoError(t, err)
	defer channel.Stop(ctx)

	// Aspetta che il server si avvii
	time.Sleep(100 * time.Millisecond)

	// 4. Esegui richiesta HTTP all'endpoint /api/analytics
	url := fmt.Sprintf("http://%s:%d/api/analytics", cfg.Host, cfg.Port)
	resp, err := http.Get(url)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// 5. Verifica risposta
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	var result []map[string]interface{}
	err = json.Unmarshal(body, &result)
	assert.NoError(t, err)

	assert.Len(t, result, 1)
	assert.Equal(t, "2026-02-19", result[0]["date"])
	assert.Equal(t, float64(262), result[0]["facebook_followers"])
	assert.Equal(t, float64(7), result[0]["linkedin_followers"])
}
