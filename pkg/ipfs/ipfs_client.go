package ipfs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents a connection to an IPFS node with configurable network settings.
type Client struct {
	scheme   string // Protocol scheme (http/https)
	host     string // Node hostname or IP
	port     int    // Node port number
	baseURL  string // Precomputed base URL for requests
	timeout  time.Duration // HTTP request timeout
}

// NewClient creates a configured IPFS client instance.
// scheme: Protocol (http/https)
// host: Node hostname/IP
// port: Node port
// timeout: HTTP request timeout (default 20s if zero)
func NewClient(scheme, host string, port int, timeout time.Duration) *Client {
	if timeout == 0 {
		timeout = 20 * time.Second
	}

	return &Client{
		scheme:  scheme,
		host:    host,
		port:    port,
		baseURL: fmt.Sprintf("%s://%s:%d/ipfs", scheme, host, port),
		timeout: timeout,
	}
}

// URLForCID constructs the full retrieval URL for a given CID.
func (c *Client) URLForCID(cid string) string {
	return fmt.Sprintf("%s/%s", c.baseURL, cid)
}

// Retrieve fetches content from IPFS by CID.
// Returns the content bytes or an error if the request fails.
func (c *Client) Retrieve(cid string) ([]byte, error) {
	return c.retrieveWithContext(context.Background(), cid)
}

// retrieveWithContext handles the actual HTTP request with context support.
func (c *Client) retrieveWithContext(ctx context.Context, cid string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.URLForCID(cid), nil)
	if err != nil {
		return nil, fmt.Errorf("request creation failed: %w", err)
	}
	req.Header.Set("Accept", "*/*")

	client := &http.Client{Timeout: c.timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}

	return data, nil
}
