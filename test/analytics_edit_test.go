package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/channels"
	"github.com/sipeed/picoclaw/pkg/config"
)

func TestAnalyticsEdit(t *testing.T) {
	// Create temp workspace
	tempDir, err := os.MkdirTemp("", "picoclaw_test_analytics_edit")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create initial social.json
	initialData := `[{"date":"2024-01-01","facebook_followers":100}]`
	socialPath := filepath.Join(tempDir, "social.json")
	err = os.WriteFile(socialPath, []byte(initialData), 0644)
	if err != nil {
		t.Fatalf("Failed to write social.json: %v", err)
	}

	// Init WebChatChannel
	msgBus := bus.NewMessageBus()
	cfg := config.WebChatConfig{
		Enabled: true,
		Port:    8081,
	}

	wc, err := channels.NewWebChatChannel(cfg, msgBus, tempDir, nil)
	if err != nil {
		t.Fatalf("Failed to create WebChatChannel: %v", err)
	}

	// Get Mux
	mux := wc.SetupMux()

	// Test GET
	reqGet := httptest.NewRequest("GET", "/api/analytics", nil)
	wGet := httptest.NewRecorder()
	mux.ServeHTTP(wGet, reqGet)

	if wGet.Code != http.StatusOK {
		t.Errorf("Expected 200 OK for GET, got %d", wGet.Code)
	}

	// Test POST (Update existing)
	updateData := channels.SocialStats{
		Date:              "2024-01-01",
		FacebookFollowers: 150,
	}
	body, _ := json.Marshal(updateData)
	reqPost := httptest.NewRequest("POST", "/api/analytics", bytes.NewBuffer(body))
	wPost := httptest.NewRecorder()
	mux.ServeHTTP(wPost, reqPost)

	if wPost.Code != http.StatusOK {
		t.Errorf("Expected 200 OK for POST, got %d. Body: %s", wPost.Code, wPost.Body.String())
	}

	// Verify Update
	data, _ := os.ReadFile(socialPath)
	var stats []channels.SocialStats
	json.Unmarshal(data, &stats)

	if len(stats) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(stats))
	}
	if len(stats) > 0 && stats[0].FacebookFollowers != 150 {
		t.Errorf("Expected 150 Facebook followers, got %d", stats[0].FacebookFollowers)
	}

	// Test POST (New Entry)
	newData := channels.SocialStats{
		Date:              "2024-01-02",
		FacebookFollowers: 200,
	}
	body, _ = json.Marshal(newData)
	reqPost2 := httptest.NewRequest("POST", "/api/analytics", bytes.NewBuffer(body))
	wPost2 := httptest.NewRecorder()
	mux.ServeHTTP(wPost2, reqPost2)

	if wPost2.Code != http.StatusOK {
		t.Errorf("Expected 200 OK for POST 2, got %d", wPost2.Code)
	}

	// Verify Append
	data, _ = os.ReadFile(socialPath)
	json.Unmarshal(data, &stats)

	if len(stats) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(stats))
	}
}
