package cms

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
)

func (cm *ContentManager) loadPrompts() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	data, err := os.ReadFile(cm.promptsPath)
	if err != nil {
		return err
	}

	var prompts []*PromptTemplate
	if err := json.Unmarshal(data, &prompts); err != nil {
		return err
	}

	cm.prompts = make(map[string]*PromptTemplate)
	for _, p := range prompts {
		cm.prompts[p.ID] = p
	}
	return nil
}

func (cm *ContentManager) savePrompts() error {
	prompts := make([]*PromptTemplate, 0, len(cm.prompts))
	for _, p := range cm.prompts {
		prompts = append(prompts, p)
	}

	data, err := json.MarshalIndent(prompts, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cm.promptsPath, data, 0644)
}

func (cm *ContentManager) ListPrompts() []*PromptTemplate {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	prompts := make([]*PromptTemplate, 0, len(cm.prompts))
	for _, p := range cm.prompts {
		prompts = append(prompts, p)
	}
	return prompts
}

func (cm *ContentManager) GetPrompt(id string) (*PromptTemplate, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if p, exists := cm.prompts[id]; exists {
		return p, nil
	}
	return nil, fmt.Errorf("prompt not found")
}

func (cm *ContentManager) CreatePrompt(name, category, description, content string) (*PromptTemplate, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	prompt := &PromptTemplate{
		ID:          uuid.New().String(),
		Name:        name,
		Category:    category,
		Description: description,
		Content:     content,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	cm.prompts[prompt.ID] = prompt
	if err := cm.savePrompts(); err != nil {
		return nil, err
	}
	return prompt, nil
}

func (cm *ContentManager) UpdatePrompt(id, name, category, description, content string) (*PromptTemplate, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	prompt, exists := cm.prompts[id]
	if !exists {
		return nil, fmt.Errorf("prompt not found")
	}

	prompt.Name = name
	prompt.Category = category
	prompt.Description = description
	prompt.Content = content
	prompt.UpdatedAt = time.Now()

	if err := cm.savePrompts(); err != nil {
		return nil, err
	}
	return prompt, nil
}

func (cm *ContentManager) DeletePrompt(id string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.prompts[id]; !exists {
		return fmt.Errorf("prompt not found")
	}

	delete(cm.prompts, id)
	return cm.savePrompts()
}

func (cm *ContentManager) ImportPrompts(prompts []*PromptTemplate) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for _, p := range prompts {
		if p.ID == "" {
			p.ID = uuid.New().String()
		}
		if p.CreatedAt.IsZero() {
			p.CreatedAt = time.Now()
		}
		p.UpdatedAt = time.Now()
		cm.prompts[p.ID] = p
	}
	return cm.savePrompts()
}
