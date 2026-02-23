package cms

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sipeed/picoclaw/pkg/logger"
	"github.com/sipeed/picoclaw/pkg/providers"
)

type ContentManager struct {
	workspace   string
	mu          sync.RWMutex
	items       map[string]*ContentItem
	prompts     map[string]*PromptTemplate
	filePath    string
	promptsPath string
	llmProvider providers.LLMProvider
}

func NewContentManager(workspace string) (*ContentManager, error) {
	cm := &ContentManager{
		workspace:   workspace,
		items:       make(map[string]*ContentItem),
		prompts:     make(map[string]*PromptTemplate),
		filePath:    filepath.Join(workspace, "data", "linkedin_content.json"),
		promptsPath: filepath.Join(workspace, "prompt-social.json"),
	}

	// Ensure data directory exists
	dataDir := filepath.Dir(cm.filePath)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Ensure prompts directory exists
	promptsDir := filepath.Dir(cm.promptsPath)
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create prompts directory: %w", err)
	}

	if err := cm.load(); err != nil {
		if !os.IsNotExist(err) {
			logger.ErrorCF("cms", "Failed to load content data", map[string]interface{}{"error": err})
		}
	}

	if err := cm.loadPrompts(); err != nil {
		if !os.IsNotExist(err) {
			logger.ErrorCF("cms", "Failed to load prompts data", map[string]interface{}{"error": err})
		}
	}

	return cm, nil
}

func (cm *ContentManager) SetLLMProvider(provider providers.LLMProvider) {
	cm.llmProvider = provider
}

func (cm *ContentManager) load() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	data, err := os.ReadFile(cm.filePath)
	if err != nil {
		return err
	}

	var items []*ContentItem
	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}

	cm.items = make(map[string]*ContentItem)
	for _, item := range items {
		cm.items[item.ID] = item
	}
	return nil
}

func (cm *ContentManager) save() error {
	items := make([]*ContentItem, 0, len(cm.items))
	for _, item := range cm.items {
		items = append(items, item)
	}

	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cm.filePath, data, 0644)
}

func (cm *ContentManager) CreateItem(req GenerateRequest) (*ContentItem, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	item := &ContentItem{
		ID:        uuid.New().String(),
		URL:       req.URL,
		Topic:     req.Topic,
		Summary:   req.Summary,
		Status:    StatusDraft,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		History:   []Revision{},
		Comments:  []Comment{},
	}

	cm.items[item.ID] = item
	if err := cm.save(); err != nil {
		return nil, err
	}
	return item, nil
}

func (cm *ContentManager) AnalyzeAndCreate(input string) (*ContentItem, error) {
	var content string
	var url string
	if strings.HasPrefix(input, "http") {
		url = input
		var err error
		content, err = cm.ScrapeURL(url)
		if err != nil {
			return nil, err
		}
	} else {
		content = input
	}

	topic, summary, err := cm.AnalyzeContent(content)
	if err != nil {
		return nil, err
	}

	return cm.CreateItem(GenerateRequest{
		URL:     url,
		Topic:   topic,
		Summary: summary,
	})
}

func (cm *ContentManager) Update(req UpdateRequest) (*ContentItem, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	item, exists := cm.items[req.ID]
	if !exists {
		return nil, fmt.Errorf("item not found")
	}

	// Save history if post text changed
	if req.PostText != "" && req.PostText != item.PostText {
		item.History = append(item.History, Revision{
			PostText:  item.PostText,
			Timestamp: time.Now(),
			Author:    "User", // TODO: Get actual user
		})
		item.PostText = req.PostText
	}

	if req.Topic != "" {
		item.Topic = req.Topic
	}

	if req.Summary != "" {
		item.Summary = req.Summary
	}

	if req.Status != "" {
		item.Status = req.Status
	}

	if req.Comment != "" {
		item.Comments = append(item.Comments, Comment{
			Author:    "Reviewer", // TODO: Get actual user
			Text:      req.Comment,
			Timestamp: time.Now(),
		})
	}

	if req.MediaURL != "" {
		item.MediaURL = req.MediaURL
	}
	if req.MediaPath != "" {
		item.MediaPath = req.MediaPath
	}
	if req.MediaType != "" {
		item.MediaType = req.MediaType
	}

	item.UpdatedAt = time.Now()
	if err := cm.save(); err != nil {
		return nil, err
	}
	return item, nil
}

func (cm *ContentManager) Get(id string) (*ContentItem, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	item, exists := cm.items[id]
	if !exists {
		return nil, fmt.Errorf("item not found")
	}
	return item, nil
}

func (cm *ContentManager) Delete(id string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.items[id]; !exists {
		return fmt.Errorf("item not found")
	}

	delete(cm.items, id)
	return cm.save()
}

func (cm *ContentManager) List() []*ContentItem {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	items := make([]*ContentItem, 0, len(cm.items))
	for _, item := range cm.items {
		items = append(items, item)
	}
	// Sort by CreatedAt desc
	// (Skipping sort for brevity, frontend can sort)
	return items
}

// ScrapeURL fetches content from a URL (simple version)
func (cm *ContentManager) ScrapeURL(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to fetch URL: status %d", resp.StatusCode)
	}

	// Read full body (limit to 1MB to prevent memory issues)
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return "", err
	}

	html := string(body)

	// Remove scripts
	reScript := regexp.MustCompile(`(?s)<script.*?>.*?</script>`)
	html = reScript.ReplaceAllString(html, "")

	// Remove styles
	reStyle := regexp.MustCompile(`(?s)<style.*?>.*?</style>`)
	html = reStyle.ReplaceAllString(html, "")

	// Remove head content if possible (though some sites put content there, unlikely)
	reHead := regexp.MustCompile(`(?s)<head.*?>.*?</head>`)
	html = reHead.ReplaceAllString(html, "")

	// Remove comments
	reComment := regexp.MustCompile(`(?s)<!--.*?-->`)
	html = reComment.ReplaceAllString(html, "")

	// Remove remaining HTML tags
	reTags := regexp.MustCompile(`<[^>]*>`)
	text := reTags.ReplaceAllString(html, " ")

	// Normalize whitespace
	reSpace := regexp.MustCompile(`\s+`)
	text = strings.TrimSpace(reSpace.ReplaceAllString(text, " "))

	// Truncate if still too long for LLM context window
	if len(text) > 15000 {
		text = text[:15000]
	}

	return text, nil
}

func (cm *ContentManager) AnalyzeContent(content string) (string, string, error) {
	if cm.llmProvider == nil {
		return "", "", fmt.Errorf("LLM provider not configured")
	}

	prompt := fmt.Sprintf(`Analyze the following text from a webpage and extract the main topic and a brief summary suitable for a LinkedIn post.
Ignore any remaining HTML code, scripts, navigation menus, or footer text. Focus on the main article content.
Return the result as a valid JSON object with keys "topic" and "summary".
Do not include any markdown formatting or explanations.

Text:
%s`, content)

	resp, err := cm.llmProvider.Chat(context.Background(), []providers.Message{
		{Role: "user", Content: prompt},
	}, nil, cm.llmProvider.GetDefaultModel(), nil)

	if err != nil {
		return "", "", err
	}

	// Clean up potential markdown code blocks
	jsonStr := resp.Content
	start := strings.Index(jsonStr, "{")
	end := strings.LastIndex(jsonStr, "}")
	if start != -1 && end != -1 && end > start {
		jsonStr = jsonStr[start : end+1]
	}

	var result struct {
		Topic   string `json:"topic"`
		Summary string `json:"summary"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return "", "", fmt.Errorf("failed to parse LLM response: %w (response: %s)", err, jsonStr)
	}

	return result.Topic, result.Summary, nil
}

func (cm *ContentManager) GeneratePostContent(topic, summary, promptTemplate string) (string, error) {
	if cm.llmProvider == nil {
		return "", fmt.Errorf("LLM provider not configured")
	}

	var prompt string
	if promptTemplate != "" {
		// Replace placeholders in custom template
		prompt = promptTemplate
		prompt = strings.ReplaceAll(prompt, "{topic}", topic)
		prompt = strings.ReplaceAll(prompt, "{summary}", summary)
	} else {
		prompt = fmt.Sprintf(`Create a professional LinkedIn post about '%s'.
Summary: %s

Requirements:
1. Length: 1300-2000 characters.
2. Tone: Professional, engaging, and informative.
3. Formatting: Use emojis and bullet points for readability. DO NOT use markdown bold (**) or italic (*) formatting. Use CAPITALIZATION or emojis for emphasis instead.
4. Hashtags: Include 3-5 relevant hashtags at the end.
5. CTA: Include a clear Call to Action.
6. Language: Italian.

Return ONLY the post text. Do not include any introductory or concluding remarks.`, topic, summary)
	}

	// Append strict instruction to ensure no chatty response
	prompt += "\n\nIMPORTANT: Return ONLY the final post text. Do not include phrases like 'Here is a draft' or 'Ecco una bozza'. DO NOT use markdown formatting like **bold** or *italic*. Use emojis for structure."

	resp, err := cm.llmProvider.Chat(context.Background(), []providers.Message{
		{Role: "user", Content: prompt},
	}, nil, cm.llmProvider.GetDefaultModel(), nil)

	if err != nil {
		return "", err
	}

	content := resp.Content

	// Clean up common prefixes if LLM is chatty
	prefixes := []string{
		"Ecco una bozza di post LinkedIn che puoi adattare e pubblicare:",
		"Ecco una bozza di post LinkedIn:",
		"Ecco il post:",
		"Here is the post:",
		"Sure, here is the post:",
		"Certainly! Here is a draft:",
	}

	for _, prefix := range prefixes {
		if idx := strings.Index(content, prefix); idx != -1 {
			// If prefix is found, take everything after it
			content = content[idx+len(prefix):]
		}
	}

	// Remove any remaining markdown asterisks
	content = strings.ReplaceAll(content, "**", "")
	content = strings.ReplaceAll(content, "*", "")

	return strings.TrimSpace(content), nil
}

func (cm *ContentManager) GeneratePostForItem(id string, promptID string) (*ContentItem, error) {
	item, err := cm.Get(id)
	if err != nil {
		return nil, err
	}

	var promptTemplate string
	if promptID != "" {
		p, err := cm.GetPrompt(promptID)
		if err != nil {
			return nil, fmt.Errorf("failed to get prompt template: %w", err)
		}
		promptTemplate = p.Content
	}

	postText, err := cm.GeneratePostContent(item.Topic, item.Summary, promptTemplate)
	if err != nil {
		return nil, err
	}

	return cm.Update(UpdateRequest{
		ID:       id,
		PostText: postText,
		Status:   StatusDraft,
	})
}
