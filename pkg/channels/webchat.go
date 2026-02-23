package channels

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/cms"
	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/logger"
	"github.com/sipeed/picoclaw/pkg/providers"
)

type WebChatChannel struct {
	BaseChannel
	config    config.WebChatConfig
	workspace string
	provider  providers.LLMProvider
	cms       *cms.ContentManager
	server    *http.Server
	wsClients map[*websocket.Conn]struct{}
	wsMu      sync.Mutex
	upgrader  websocket.Upgrader
}

type SocialStats struct {
	Date               string `json:"date"`
	FacebookFollowers  int    `json:"facebook_followers"`
	FacebookPosts      int    `json:"facebook_posts"`
	InstagramFollowers int    `json:"instagram_followers"`
	InstagramPosts     int    `json:"instagram_posts"`
	LinkedinFollowers  int    `json:"linkedin_followers"`
	TwitterFollowers   int    `json:"twitter_followers"`
	TwitterPosts       int    `json:"twitter_posts"`
	YouTubeSubscribers int    `json:"youtube_subscribers"`
	YouTubeVideos      int    `json:"youtube_videos"`
}

func NewWebChatChannel(cfg config.WebChatConfig, messageBus *bus.MessageBus, workspace string, provider providers.LLMProvider) (*WebChatChannel, error) {
	cmsManager, err := cms.NewContentManager(workspace)
	if err != nil {
		return nil, err
	} else {
		cmsManager.SetLLMProvider(provider)
	}

	base := NewBaseChannel("webchat", cfg, messageBus, cfg.AllowFrom)

	return &WebChatChannel{
		BaseChannel: *base,
		config:      cfg,
		workspace:   workspace,
		provider:    provider,
		cms:         cmsManager,
		wsClients:   make(map[*websocket.Conn]struct{}),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}, nil
}

func (c *WebChatChannel) Start(ctx context.Context) error {
	host := c.config.Host
	if host == "" {
		host = "0.0.0.0"
	}
	port := c.config.Port
	if port == 0 {
		port = 8000
	}
	mux := c.SetupMux()

	c.server = &http.Server{Addr: fmt.Sprintf("%s:%d", host, port), Handler: mux}

	go func() {
		logger.InfoCF("channels", "WebChat channel starting", map[string]interface{}{"port": port})
		if err := c.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.ErrorCF("channels", "WebChat server error", map[string]interface{}{"error": err})
		}
	}()

	c.running = true
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
	c.wsMu.Lock()
	clients := make([]*websocket.Conn, 0, len(c.wsClients))
	for conn := range c.wsClients {
		clients = append(clients, conn)
	}
	c.wsMu.Unlock()

	for _, conn := range clients {
		if err := conn.WriteJSON(map[string]interface{}{
			"type":    "message",
			"content": msg.Content,
		}); err != nil {
			c.removeClient(conn)
		}
	}
	return nil
}

func (c *WebChatChannel) SetupMux() *http.ServeMux {
	mux := http.NewServeMux()

	uploadsDir := filepath.Join(c.workspace, "data", "uploads")
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		logger.ErrorCF("channels", "Failed to create uploads dir", map[string]interface{}{"error": err})
	}
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadsDir))))

	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := c.upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		c.addClient(conn)
		defer func() {
			c.removeClient(conn)
			conn.Close()
		}()
		for {
			var payload struct {
				Content string `json:"content"`
			}
			if err := conn.ReadJSON(&payload); err != nil {
				return
			}
			if strings.TrimSpace(payload.Content) == "" {
				continue
			}
			c.HandleMessage("web_user", "webchat", payload.Content, nil, nil)
		}
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "index.html")
	})

	mux.HandleFunc("/crypto", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "crypto.html")
	})

	mux.HandleFunc("/linkedin", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "linkedin.html")
	})

	mux.HandleFunc("/linkedin/cms", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "linkedin_cms.html")
	})

	mux.HandleFunc("/api/cms/list", func(w http.ResponseWriter, r *http.Request) {
		items := c.cms.List()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(items)
	})

	mux.HandleFunc("/api/cms/get", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		item, err := c.cms.Get(id)
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(item)
	})

	mux.HandleFunc("/api/cms/analyze", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Input string `json:"input"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		item, err := c.cms.AnalyzeAndCreate(req.Input)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(item)
	})

	mux.HandleFunc("/api/cms/generate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			ID       string `json:"id"`
			PromptID string `json:"prompt_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		item, err := c.cms.GeneratePostForItem(req.ID, req.PromptID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(item)
	})

	mux.HandleFunc("/api/cms/prompts", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(c.cms.ListPrompts())
		case http.MethodPost:
			var req struct {
				Name        string `json:"name"`
				Category    string `json:"category"`
				Description string `json:"description"`
				Content     string `json:"content"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			prompt, err := c.cms.CreatePrompt(req.Name, req.Category, req.Description, req.Content)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			json.NewEncoder(w).Encode(prompt)
		case http.MethodPut:
			var req struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				Category    string `json:"category"`
				Description string `json:"description"`
				Content     string `json:"content"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			prompt, err := c.cms.UpdatePrompt(req.ID, req.Name, req.Category, req.Description, req.Content)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			json.NewEncoder(w).Encode(prompt)
		case http.MethodDelete:
			id := r.URL.Query().Get("id")
			if id == "" {
				http.Error(w, "missing id", 400)
				return
			}
			if err := c.cms.DeletePrompt(id); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/cms/prompts/export", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=\"prompts.json\"")
		json.NewEncoder(w).Encode(c.cms.ListPrompts())
	})

	mux.HandleFunc("/api/cms/prompts/import", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var prompts []*cms.PromptTemplate
		if err := json.NewDecoder(r.Body).Decode(&prompts); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if err := c.cms.ImportPrompts(prompts); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "imported"})
	})

	mux.HandleFunc("/api/cms/update", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req cms.UpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		item, err := c.cms.Update(req)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(item)
	})

	mux.HandleFunc("/api/cms/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		if err := c.cms.Delete(req.ID); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
	})

	mux.HandleFunc("/api/chat", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Chat endpoint placeholder"))
	})

	mux.HandleFunc("/api/linkedin/auth", func(w http.ResponseWriter, r *http.Request) {
		authURL := "https://www.linkedin.com/oauth/v2/authorization?response_type=code&client_id=" + os.Getenv("LINKEDIN_CLIENT_ID") + "&redirect_uri=" + os.Getenv("LINKEDIN_REDIRECT_URI") + "&scope=openid%20profile%20w_member_social%20email"
		http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
	})

	mux.HandleFunc("/api/linkedin/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Code string `json:"code"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		scriptPath := filepath.Join(c.workspace, "..", "get_linkedin_token.py")
		cmd := exec.Command("python3", scriptPath, "--code", req.Code)
		cmd.Dir = filepath.Join(c.workspace, "..")

		output, err := cmd.CombinedOutput()
		if err != nil {
			logger.ErrorCF("channels", "Failed to exchange token", map[string]interface{}{"error": err.Error(), "output": string(output)})
			http.Error(w, fmt.Sprintf("Error: %v, Output: %s", err, output), 500)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		var tokenResp map[string]interface{}
		if err := json.Unmarshal(output, &tokenResp); err == nil {
			if accessToken, ok := tokenResp["access_token"].(string); ok {
				envPath := filepath.Join(c.workspace, "..", ".env")
				content, err := os.ReadFile(envPath)
				if err == nil {
					lines := strings.Split(string(content), "\n")
					var newLines []string
					found := false
					for _, line := range lines {
						if strings.HasPrefix(line, "LINKEDIN_ACCESS_TOKEN=") {
							newLines = append(newLines, "LINKEDIN_ACCESS_TOKEN="+accessToken)
							found = true
						} else {
							newLines = append(newLines, line)
						}
					}
					if !found {
						newLines = append(newLines, "LINKEDIN_ACCESS_TOKEN="+accessToken)
					}
					os.WriteFile(envPath, []byte(strings.Join(newLines, "\n")), 0644)
				}
			}
		}

		w.Write(output)
	})

	mux.HandleFunc("/api/linkedin/companies", func(w http.ResponseWriter, r *http.Request) {
		scriptPath := filepath.Join(c.workspace, "..", "post_to_linkedin.py")
		cmd := exec.Command("python3", scriptPath, "--list-companies", "--json")
		cmd.Dir = filepath.Join(c.workspace, "..")

		output, err := cmd.CombinedOutput()
		if err != nil {
			logger.ErrorCF("channels", "Failed to list companies", map[string]interface{}{"error": err.Error(), "output": string(output)})
			http.Error(w, fmt.Sprintf("Error: %v, Output: %s", err, output), 500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(output)
	})

	mux.HandleFunc("/api/linkedin/post", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Text      string `json:"text"`
			MediaPath string `json:"media_path"`
			MediaType string `json:"media_type"`
			AuthorURN string `json:"author_urn"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		scriptPath := filepath.Join(c.workspace, "..", "post_to_linkedin.py")
		args := []string{scriptPath, req.Text, "--json"}
		if req.MediaPath != "" {
			args = append(args, "--media", req.MediaPath)
		}
		if req.AuthorURN != "" {
			args = append(args, "--author", req.AuthorURN)
		}

		cmd := exec.Command("python3", args...)
		cmd.Dir = filepath.Join(c.workspace, "..")

		output, err := cmd.CombinedOutput()
		if err != nil {
			logger.ErrorCF("channels", "Failed to post", map[string]interface{}{"error": err.Error(), "output": string(output)})
			http.Error(w, fmt.Sprintf("Error: %v, Output: %s", err, output), 500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(output)
	})

	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	mux.HandleFunc("/api/analytics", func(w http.ResponseWriter, r *http.Request) {
		statsPath := filepath.Join(c.workspace, "social.json")
		switch r.Method {
		case http.MethodGet:
			data, err := os.ReadFile(statsPath)
			if err != nil {
				if os.IsNotExist(err) {
					w.Header().Set("Content-Type", "application/json")
					w.Write([]byte("[]"))
					return
				}
				http.Error(w, "Failed to load analytics", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(data)
		case http.MethodPost:
			var payload SocialStats
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			if strings.TrimSpace(payload.Date) == "" {
				http.Error(w, "missing date", 400)
				return
			}
			var stats []SocialStats
			if data, err := os.ReadFile(statsPath); err == nil {
				json.Unmarshal(data, &stats)
			}
			updated := false
			for i := range stats {
				if stats[i].Date == payload.Date {
					stats[i] = payload
					updated = true
					break
				}
			}
			if !updated {
				stats = append(stats, payload)
			}
			data, err := json.MarshalIndent(stats, "", "  ")
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			if err := os.WriteFile(statsPath, data, 0644); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(data)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

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

	mux.HandleFunc("/api/news/refresh", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		scriptPath := filepath.Join(c.workspace, "news", "fetch_rss_news.py")
		cmd := exec.Command("python3", scriptPath)
		if output, err := cmd.CombinedOutput(); err != nil {
			logger.ErrorCF("channels", "Failed to refresh news", map[string]interface{}{
				"error":  err.Error(),
				"output": string(output),
			})
			http.Error(w, fmt.Sprintf("Failed to refresh news: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	mux.HandleFunc("/api/weather", func(w http.ResponseWriter, r *http.Request) {
		city := r.URL.Query().Get("city")
		scriptPath := filepath.Join(c.workspace, "skills", "weather", "weather.py")
		var cmd *exec.Cmd
		if city != "" {
			cmd = exec.Command("python3", scriptPath, city)
		} else {
			cmd = exec.Command("python3", scriptPath)
		}

		output, err := cmd.CombinedOutput()
		if err != nil {
			logger.ErrorCF("channels", "Failed to fetch weather", map[string]interface{}{
				"error":  err.Error(),
				"output": string(output),
			})
			http.Error(w, fmt.Sprintf("Failed to fetch weather: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Write(output)
	})

	mux.HandleFunc("/api/crypto", func(w http.ResponseWriter, r *http.Request) {
		cryptoPath := filepath.Join(c.workspace, "news", "crypto.json")
		data, err := os.ReadFile(cryptoPath)
		if err != nil {
			logger.ErrorCF("channels", "Failed to read crypto file", map[string]interface{}{
				"path":  cryptoPath,
				"error": err.Error(),
			})
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("[]"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	})

	mux.HandleFunc("/api/crypto/refresh", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		scriptPath := filepath.Join(c.workspace, "news", "fetch_crypto.py")
		cmd := exec.Command("python3", scriptPath)
		if output, err := cmd.CombinedOutput(); err != nil {
			logger.ErrorCF("channels", "Failed to refresh crypto", map[string]interface{}{
				"error":  err.Error(),
				"output": string(output),
			})
			http.Error(w, fmt.Sprintf("Failed to refresh crypto: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	mux.HandleFunc("/api/cms/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseMultipartForm(50 << 20); err != nil {
			http.Error(w, "File too large", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Invalid file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		ext := filepath.Ext(header.Filename)
		filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
		destPath := filepath.Join(uploadsDir, filename)

		dst, err := os.Create(destPath)
		if err != nil {
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, "Failed to write file", http.StatusInternalServerError)
			return
		}

		mediaType := header.Header.Get("Content-Type")

		resp := map[string]string{
			"url":  fmt.Sprintf("/uploads/%s", filename),
			"path": destPath,
			"type": mediaType,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	return mux
}

func (c *WebChatChannel) addClient(conn *websocket.Conn) {
	c.wsMu.Lock()
	c.wsClients[conn] = struct{}{}
	c.wsMu.Unlock()
}

func (c *WebChatChannel) removeClient(conn *websocket.Conn) {
	c.wsMu.Lock()
	delete(c.wsClients, conn)
	c.wsMu.Unlock()
}
