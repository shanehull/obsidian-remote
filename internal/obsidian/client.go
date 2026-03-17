package obsidian

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/shanehull/obsidian-remote/internal/config"
)

type Client struct {
	cfg  *config.Config
	http *http.Client
}

func NewClient(cfg *config.Config) *Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // internal loopback only
	}
	return &Client{
		cfg:  cfg,
		http: &http.Client{Transport: tr},
	}
}

func (c *Client) Call(method, path string, body []byte, headers ...map[string]string) ([]byte, error) {
	req, err := http.NewRequest(method, c.cfg.ObsidianURL+path, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.cfg.ObsidianKey)
	req.Header.Set("Content-Type", contentTypeFor(method, path))

	for _, h := range headers {
		for k, v := range h {
			req.Header.Set(k, v)
		}
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("obsidian API %s %s returned %d: %s", method, path, resp.StatusCode, respBody)
	}

	return respBody, nil
}

// contentTypeFor returns the appropriate Content-Type for an Obsidian REST API request.
// Vault write operations expect text/markdown; everything else uses application/json.
func contentTypeFor(method, path string) string {
	if strings.HasPrefix(path, "/vault/") && (method == http.MethodPut || method == http.MethodPost) {
		return "text/markdown"
	}
	return "application/json"
}
