package supabase

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient() (*Client, error) {
	baseURL := strings.TrimRight(os.Getenv("SUPABASE_URL"), "/")
	apiKey := os.Getenv("SUPABASE_KEY")
	if baseURL == "" || apiKey == "" {
		return nil, fmt.Errorf("SUPABASE_URL or SUPABASE_KEY is missing")
	}

	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

func (client *Client) Do(ctx context.Context, method string, endpoint string, body io.Reader, prefer string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, client.baseURL+"/rest/v1/"+endpoint, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("apikey", client.apiKey)
	req.Header.Set("Authorization", "Bearer "+client.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if prefer != "" {
		req.Header.Set("Prefer", prefer)
	}

	return client.httpClient.Do(req)
}
