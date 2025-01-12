package rest

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/coding-hui/ai-terminal/internal/errbook"
)

const (
	defaultTimeout = 30 * time.Second
)

// IsRemoteFile checks if a path is a remote URL
func IsRemoteFile(path string) bool {
	u, err := url.Parse(path)
	if err != nil {
		return false
	}
	return u.Scheme != "" && (u.Scheme == "http" || u.Scheme == "https")
}

// FetchRemoteContent downloads content from a remote URL with timeout
func FetchRemoteContent(url string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", errbook.Wrap("Failed to create request", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", errbook.Wrap("Failed to fetch remote content", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errbook.Wrap("Failed to read response body", err)
	}

	return string(content), nil
}
