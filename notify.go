package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Notifier struct {
	store *ConfigStore
	http  *http.Client
}

func NewNotifier(store *ConfigStore) *Notifier {
	return &Notifier{store: store, http: &http.Client{Timeout: 15 * time.Second}}
}

// Send publishes a message via ntfy. Titles must stay ASCII (HTTP header);
// emoji go through Tags, which ntfy renders in front of the title.
func (n *Notifier) Send(title, message, priority, tags string) error {
	cfg := n.store.Get()
	if cfg.NtfyTopic == "" {
		return errors.New("ntfy topic is not configured")
	}
	server := strings.TrimRight(cfg.NtfyServer, "/")
	if server == "" {
		server = "https://ntfy.sh"
	}
	req, err := http.NewRequest(http.MethodPost, server+"/"+url.PathEscape(cfg.NtfyTopic), strings.NewReader(message))
	if err != nil {
		return err
	}
	req.Header.Set("Title", title)
	req.Header.Set("Priority", priority)
	req.Header.Set("Tags", tags)
	if cfg.NtfyToken != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.NtfyToken)
	}
	resp, err := n.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("ntfy: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return nil
}
