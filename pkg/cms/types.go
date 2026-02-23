package cms

import (
	"time"
)

type ContentStatus string

const (
	StatusDraft     ContentStatus = "DRAFT"
	StatusPending   ContentStatus = "PENDING_APPROVAL"
	StatusApproved  ContentStatus = "APPROVED"
	StatusRejected  ContentStatus = "REJECTED"
	StatusPublished ContentStatus = "PUBLISHED"
)

type ContentItem struct {
	ID        string        `json:"id"`
	URL       string        `json:"url,omitempty"`
	Topic     string        `json:"topic"`
	Summary   string        `json:"summary"`
	Prompt    string        `json:"prompt"`
	PostText  string        `json:"post_text"`
	Status    ContentStatus `json:"status"`
	Comments  []Comment     `json:"comments"`
	History   []Revision    `json:"history"`
	MediaURL  string        `json:"media_url,omitempty"`
	MediaPath string        `json:"media_path,omitempty"`
	MediaType string        `json:"media_type,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type Comment struct {
	Author    string    `json:"author"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}

type Revision struct {
	PostText  string    `json:"post_text"`
	Timestamp time.Time `json:"timestamp"`
	Author    string    `json:"author"`
}

type GenerateRequest struct {
	URL     string `json:"url"`
	Topic   string `json:"topic"`
	Summary string `json:"summary"`
}

type UpdateRequest struct {
	ID        string        `json:"id"`
	PostText  string        `json:"post_text"`
	Topic     string        `json:"topic"`
	Summary   string        `json:"summary"`
	Status    ContentStatus `json:"status"`
	Comment   string        `json:"comment"`
	MediaURL  string        `json:"media_url"`
	MediaPath string        `json:"media_path"`
	MediaType string        `json:"media_type"`
}

type PromptTemplate struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
