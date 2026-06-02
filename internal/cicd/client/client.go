// internal/cicd/client/client.go

package cicdclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	cicdmodels "github.com/SAP/terraform-provider-btp-services/internal/cicd/models"
)

// cicdHTTPClient is the low-level HTTP transport.
// It is unexported — callers always go through CicdClientFacade.
type cicdHTTPClient struct {
	mu           sync.Mutex
	httpClient   *http.Client
	baseURL      string
	tokenURL     string
	clientID     string
	clientSecret string
	token        string
	tokenExpiry  time.Time
}

func newCicdHTTPClient(cfg CicdClientConfig) *cicdHTTPClient {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	return &cicdHTTPClient{
		httpClient: &http.Client{
			Timeout: timeout,
			Transport: &retryTransport{
				wrapped:    http.DefaultTransport,
				maxRetries: 3,
				retryDelay: 2 * time.Second,
			},
		},
		baseURL:      strings.TrimRight(cfg.Endpoint, "/"),
		tokenURL:     cfg.TokenURL,
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
	}
}

// newCicdHTTPClientWithHTTP creates a client with an injected *http.Client.
// Used exclusively by VCR acceptance tests.
func newCicdHTTPClientWithHTTP(cfg CicdClientConfig, httpClient *http.Client) *cicdHTTPClient {
	hc := newCicdHTTPClient(cfg)
	hc.httpClient = httpClient
	return hc
}

// bearerToken returns a valid OAuth2 access token, fetching a new one if the
// cached token is missing or within 30 seconds of expiry.
func (c *cicdHTTPClient) bearerToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token != "" && time.Now().Add(30*time.Second).Before(c.tokenExpiry) {
		return c.token, nil
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenURL,
		strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("build token request: %w", err)
	}
	req.SetBasicAuth(c.clientID, c.clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch token: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}

	c.token = result.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)
	return c.token, nil
}

func (c *cicdHTTPClient) doGet(ctx context.Context, path string, out any) error {
	token, err := c.bearerToken(ctx)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := cicdmodels.CheckAPIResponse(resp, path); err != nil {
		return err
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *cicdHTTPClient) doPost(ctx context.Context, path string, body any) error {
	return c.doWithBody(ctx, http.MethodPost, path, body, nil)
}

func (c *cicdHTTPClient) doPut(ctx context.Context, path string, body any) error {
	return c.doWithBody(ctx, http.MethodPut, path, body, nil)
}

func (c *cicdHTTPClient) doPatch(ctx context.Context, path string, body any) error {
	token, err := c.bearerToken(ctx)
	if err != nil {
		return err
	}

	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal patch request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.baseURL+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/merge-patch+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return cicdmodels.CheckAPIResponse(resp, path)
}

func (c *cicdHTTPClient) doDelete(ctx context.Context, path string) error {
	token, err := c.bearerToken(ctx)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return cicdmodels.CheckAPIResponse(resp, path)
}

func (c *cicdHTTPClient) doWithBody(ctx context.Context, method, path string, body any, out any) error {
	token, err := c.bearerToken(ctx)
	if err != nil {
		return err
	}

	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := cicdmodels.CheckAPIResponse(resp, path); err != nil {
		return err
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

// retryTransport wraps http.RoundTripper with linear-backoff retries for
// 429 (rate limit) and 5xx responses.
type retryTransport struct {
	wrapped    http.RoundTripper
	maxRetries int
	retryDelay time.Duration
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var (
		resp *http.Response
		err  error
	)
	for attempt := 0; attempt <= t.maxRetries; attempt++ {
		resp, err = t.wrapped.RoundTrip(req)
		if err != nil || !isRetryable(resp) {
			return resp, err
		}
		time.Sleep(t.retryDelay * time.Duration(attempt+1))
	}
	return resp, err
}

func isRetryable(resp *http.Response) bool {
	if resp == nil {
		return true
	}
	return resp.StatusCode == http.StatusTooManyRequests ||
		resp.StatusCode >= http.StatusInternalServerError
}
