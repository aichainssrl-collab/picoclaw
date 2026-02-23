package cms

import (
	"context"
	"os"
	"testing"

	"github.com/sipeed/picoclaw/pkg/providers"
)

// MockLLMProvider implements providers.LLMProvider for testing
type MockLLMProvider struct {
	ResponseContent string
	LastPrompt      string
}

func (m *MockLLMProvider) Chat(ctx context.Context, messages []providers.Message, tools []providers.ToolDefinition, model string, options map[string]interface{}) (*providers.LLMResponse, error) {
	if len(messages) > 0 {
		m.LastPrompt = messages[0].Content
	}
	return &providers.LLMResponse{
		Content: m.ResponseContent,
	}, nil
}

func (m *MockLLMProvider) GetDefaultModel() string {
	return "mock-model"
}

func TestContentManager_CRUD(t *testing.T) {
	// Setup temp workspace
	tmpDir, err := os.MkdirTemp("", "cms_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cm, err := NewContentManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test Create
	req := GenerateRequest{
		URL:   "http://example.com",
		Topic: "Test Topic",
	}
	item, err := cm.CreateItem(req)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if item.Topic != "Test Topic" {
		t.Errorf("Expected topic 'Test Topic', got '%s'", item.Topic)
	}
	if item.Status != StatusDraft {
		t.Errorf("Expected status 'DRAFT', got '%s'", item.Status)
	}

	// Test Get
	gotItem, err := cm.Get(item.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if gotItem.ID != item.ID {
		t.Errorf("IDs mismatch")
	}

	// Test Update
	updateReq := UpdateRequest{
		ID:       item.ID,
		PostText: "Updated text",
		Status:   StatusPending,
	}
	updatedItem, err := cm.Update(updateReq)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updatedItem.PostText != "Updated text" {
		t.Errorf("PostText not updated")
	}
	if updatedItem.Status != StatusPending {
		t.Errorf("Status not updated")
	}

	// Verify persistence
	cm2, err := NewContentManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create second manager: %v", err)
	}
	loadedItem, err := cm2.Get(item.ID)
	if err != nil {
		t.Fatalf("Failed to load item from disk: %v", err)
	}
	if loadedItem.PostText != "Updated text" {
		t.Errorf("Persistence failed, expected 'Updated text', got '%s'", loadedItem.PostText)
	}

	// Test Delete
	err = cm.Delete(item.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	_, err = cm.Get(item.ID)
	if err == nil {
		t.Errorf("Expected error when getting deleted item, got nil")
	}
}

func TestContentManager_LLM(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cms_llm_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cm, err := NewContentManager(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	mockLLM := &MockLLMProvider{}
	cm.SetLLMProvider(mockLLM)

	// Test AnalyzeContent
	mockLLM.ResponseContent = `{"topic": "AI News", "summary": "AI is growing fast."}`
	topic, summary, err := cm.AnalyzeContent("Some long text about AI...")
	if err != nil {
		t.Fatalf("AnalyzeContent failed: %v", err)
	}
	if topic != "AI News" {
		t.Errorf("Expected topic 'AI News', got '%s'", topic)
	}
	if summary != "AI is growing fast." {
		t.Errorf("Expected summary 'AI is growing fast.', got '%s'", summary)
	}

	// Test GeneratePost
	mockLLM.ResponseContent = "Here is a LinkedIn post."
	post, err := cm.GeneratePostContent("AI", "Summary", "")
	if err != nil {
		t.Fatalf("GeneratePost failed: %v", err)
	}
	if post != "Here is a LinkedIn post." {
		t.Errorf("Unexpected post content: %s", post)
	}
}
