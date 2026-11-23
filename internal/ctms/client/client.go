package ctmsclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	ctmsmodels "github.com/SAP/terraform-provider-btp-services/internal/ctms/models"
)

// ctmsHTTPClient owns transport and authentication, never Terraform state.
// Configuration is copied at construction and is immutable thereafter.
type ctmsHTTPClient struct {
	mu           sync.Mutex
	httpClient   *http.Client
	baseURL      string
	tokenURL     string
	clientID     string
	clientSecret string
	userAgent    string
	token        string
	tokenExpiry  time.Time
}

func newCtmsHTTPClient(cfg CtmsClientConfig) *ctmsHTTPClient {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	return &ctmsHTTPClient{
		httpClient: &http.Client{
			Timeout:       timeout,
			CheckRedirect: checkHTTPSRedirect,
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
		userAgent:    cfg.UserAgent,
	}
}

func newCtmsHTTPClientWithHTTP(cfg CtmsClientConfig, client *http.Client) *ctmsHTTPClient {
	hc := newCtmsHTTPClient(cfg)
	hc.httpClient = client
	return hc
}

// checkHTTPSRedirect prevents credentials from being sent over an unencrypted
// redirect and preserves net/http's default limit of ten requests.
func checkHTTPSRedirect(req *http.Request, via []*http.Request) error {
	if req.URL.Scheme != "https" {
		return fmt.Errorf("CTMS redirects must use HTTPS")
	}
	if len(via) >= 10 {
		return fmt.Errorf("stopped after 10 redirects")
	}
	return nil
}

// tokenRetryKey marks only OAuth token requests as safe to retry with POST.
type tokenRetryKey struct{}

// bearerToken caches tokens and refreshes within 30 seconds of expiry, as CI/CD does.
func (c *ctmsHTTPClient) bearerToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if c.token != "" && time.Now().Add(30*time.Second).Before(c.tokenExpiry) {
		return c.token, nil
	}
	data := url.Values{"grant_type": {"client_credentials"}}
	tokenContext := context.WithValue(ctx, tokenRetryKey{}, true)
	req, err := http.NewRequestWithContext(tokenContext, http.MethodPost, c.tokenURL, strings.NewReader(data.Encode()))
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
		if err := ctmsmodels.CheckAPIResponse(resp, c.tokenURL); err != nil {
			// Strip the resource-not-found marker from OAuth failures. Wrapping
			// it with %w would let a resource Read incorrectly remove state.
			if notFound, ok := errors.AsType[*ctmsmodels.NotFoundError](err); ok {
				err = notFound.APIError
			}
			return "", fmt.Errorf("fetch token: %w", err)
		}
		return "", fmt.Errorf("token endpoint returned unexpected status %d", resp.StatusCode)
	}
	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}
	if result.AccessToken == "" {
		return "", fmt.Errorf("token response is missing access_token")
	}
	// As in CI/CD, a missing/nonpositive lifetime prevents cache reuse.
	c.token = result.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)
	return c.token, nil
}

func (c *ctmsHTTPClient) doGet(ctx context.Context, path string, out any) error {
	return c.do(ctx, http.MethodGet, path, nil, out)
}

func (c *ctmsHTTPClient) doDelete(ctx context.Context, path string) error {
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}

func (c *ctmsHTTPClient) doWithBody(ctx context.Context, method, path string, body, out any) error {
	return c.do(ctx, method, path, body, out)
}

func (c *ctmsHTTPClient) do(ctx context.Context, method, path string, body, out any) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal %s %s: %w", method, path, err)
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return fmt.Errorf("build %s %s: %w", method, path, err)
	}
	token, err := c.bearerToken(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", c.userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%s %s: %w", method, path, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if err := ctmsmodels.CheckAPIResponse(resp, path); err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	// All landscape responses are JSON objects. Reject null instead of silently
	// returning a zero-valued object that a provider could mistake for state.
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read %s %s response: %w", method, path, err)
	}
	if trimmed := bytes.TrimSpace(b); len(trimmed) == 0 || trimmed[0] != '{' {
		return fmt.Errorf("decode %s %s response: expected JSON object", method, path)
	}
	if err := json.Unmarshal(b, out); err != nil {
		return fmt.Errorf("decode %s %s response: %w", method, path, err)
	}
	return nil
}

// retryTransport retries GET, DELETE, and explicitly marked OAuth token POSTs
// on 429/5xx. Resource POSTs and PATCHes are never automatically retried: a
// failed response does not establish that the server rejected the mutation.
type retryTransport struct {
	wrapped    http.RoundTripper
	maxRetries int
	retryDelay time.Duration
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	canRetry := req.Method == http.MethodGet || req.Method == http.MethodDelete ||
		(req.Method == http.MethodPost && req.Context().Value(tokenRetryKey{}) == true)
	current := req
	for attempt := 0; ; attempt++ {
		resp, err := t.wrapped.RoundTrip(current)
		if err != nil || !canRetry || !isRetryable(resp) || attempt >= t.maxRetries {
			return resp, err
		}
		// All client-generated bodies are replayable. Do not resend an exhausted
		// body if this transport is used with a streaming request in the future.
		if req.Body != nil && req.Body != http.NoBody && req.GetBody == nil {
			return resp, nil
		}
		_ = resp.Body.Close()
		timer := time.NewTimer(t.retryDelay * time.Duration(attempt+1))
		select {
		case <-req.Context().Done():
			timer.Stop()
			return nil, req.Context().Err()
		case <-timer.C:
		}
		if err := req.Context().Err(); err != nil {
			return nil, err
		}
		current = req.Clone(req.Context())
		if req.GetBody != nil {
			current.Body, err = req.GetBody()
			if err != nil {
				return nil, fmt.Errorf("replay request body: %w", err)
			}
		}
	}
}

func isRetryable(resp *http.Response) bool {
	return resp != nil && (resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= http.StatusInternalServerError)
}
