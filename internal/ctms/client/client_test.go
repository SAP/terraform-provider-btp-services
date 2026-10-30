package ctmsclient

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	ctmsmodels "github.com/SAP/terraform-provider-btp-services/internal/ctms/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func response(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestConstructorsAndConcurrentTokenCache(t *testing.T) {
	var tokens atomic.Int32
	var mutations atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			tokens.Add(1)
			assert.Empty(t, r.Header.Get("tms-named-user"))
			_, _ = w.Write([]byte(`{"access_token":"cached","expires_in":3600}`))
			return
		}
		mutations.Add(1)
		assert.Equal(t, "/v2/nodes/1", r.URL.Path)
		assert.Equal(t, "Bearer cached", r.Header.Get("Authorization"))
		assert.Equal(t, "terraform/test", r.Header.Get("User-Agent"))
		assert.Equal(t, "automation", r.Header.Get("tms-named-user"))
		w.WriteHeader(204)
	}))
	defer srv.Close()

	cfg := CtmsClientConfig{
		Endpoint:  srv.URL,
		TokenURL:  srv.URL + "/oauth/token",
		NamedUser: "automation",
		UserAgent: "terraform/test",
	}
	c := NewCtmsClientFacade(cfg)
	require.Equal(t, DefaultTimeout, c.httpClient.httpClient.Timeout)
	rt, ok := c.httpClient.httpClient.Transport.(*retryTransport)
	require.True(t, ok)
	require.Equal(t, 3, rt.maxRetries)
	require.Equal(t, 2*time.Second, rt.retryDelay)
	cfg.NamedUser = "modified externally"

	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			assert.NoError(t, c.Nodes.Delete(context.Background(), 1))
		})
	}
	wg.Wait()
	require.Equal(t, int32(1), tokens.Load())
	require.Equal(t, int32(20), mutations.Load())

	c.httpClient.mu.Lock()
	c.httpClient.tokenExpiry = time.Now().Add(20 * time.Second)
	c.httpClient.mu.Unlock()
	require.NoError(t, c.Nodes.Delete(context.Background(), 1))
	require.Equal(t, int32(2), tokens.Load())

	custom := &http.Client{Timeout: 7 * time.Second}
	injected := NewCtmsClientFacadeWithHTTP(cfg, custom)
	require.Same(t, custom, injected.httpClient.httpClient)
	cfg.Timeout = 5 * time.Second
	require.Equal(t, 5*time.Second, NewCtmsClientFacade(cfg).httpClient.httpClient.Timeout)
}

func TestTokenFailuresAreNotResourceNotFound(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
	}{
		{"missing endpoint", 404, `{"message":"missing token endpoint"}`},
		{"unauthorized", 401, `{"error":"unauthorized","error_description":"bad credentials"}`},
		{"invalid JSON", 200, "broken"},
		{"missing token", 200, `{"expires_in":3600}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var calls int
			transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				assert.Equal(t, "auth.test", r.URL.Host)
				return response(tc.status, tc.body), nil
			})
			cfg := CtmsClientConfig{
				Endpoint: "https://api.test",
				TokenURL: "https://auth.test/token",
			}
			c := NewCtmsClientFacadeWithHTTP(cfg, &http.Client{Transport: transport})
			_, err := c.Nodes.Get(context.Background(), 1)
			require.Error(t, err)
			require.False(t, ctmsmodels.IsNotFound(err))
			if tc.status >= http.StatusBadRequest {
				var apiErr *ctmsmodels.APIError
				require.ErrorAs(t, err, &apiErr)
				require.Equal(t, tc.status, apiErr.StatusCode)
				require.Equal(t, cfg.TokenURL, apiErr.Reference)
				require.NotEmpty(t, apiErr.Message)
				require.Contains(t, err.Error(), "fetch token:")
			}
			require.Equal(t, 1, calls)
		})
	}
}

func TestNamedUserValidationAndClientIsolation(t *testing.T) {
	invalid := []string{
		strings.Repeat("x", 61),
		"line\nfeed",
		"carriage\rreturn",
		string([]byte{0}),
		string([]byte{127}),
		string([]byte{255}),
	}
	for _, value := range invalid {
		t.Run(fmt.Sprintf("invalid-%q", value), func(t *testing.T) {
			var calls int
			transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				return response(204, ""), nil
			})
			cfg := CtmsClientConfig{Endpoint: "https://api.test", NamedUser: value}
			c := NewCtmsClientFacadeWithHTTP(cfg, &http.Client{Transport: transport})
			require.Error(t, c.Nodes.Delete(context.Background(), 1))
			require.Zero(t, calls)
		})
	}

	var wg sync.WaitGroup
	for _, value := range []string{"", strings.Repeat("ä", 60), "alias-one", "alias-two"} {
		c := prepareClientFacadeForTest(t, value, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, value, r.Header.Get("tms-named-user"))
			w.WriteHeader(204)
		})
		for range 5 {
			wg.Go(func() {
				assert.NoError(t, c.Routes.Delete(context.Background(), 1))
			})
		}
	}
	wg.Wait()
}

type trackedBody struct {
	io.Reader
	closed bool
}

func (b *trackedBody) Close() error {
	b.closed = true
	return nil
}

func TestRetryPolicyAndBodyReplay(t *testing.T) {
	for _, method := range []string{"GET", "POST", "PATCH", "DELETE"} {
		for _, status := range []int{200, 204, 400, 401, 403, 404, 409, 422, 429, 500, 503} {
			t.Run(fmt.Sprintf("%s/%d", method, status), func(t *testing.T) {
				var bodies []*trackedBody
				var calls int
				transport := &retryTransport{
					maxRetries: 3,
					wrapped: roundTripFunc(func(r *http.Request) (*http.Response, error) {
						calls++
						b, err := io.ReadAll(r.Body)
						assert.NoError(t, err)
						assert.Equal(t, `{"payload":true}`, string(b))
						_ = r.Body.Close()
						assert.Equal(t, "automation", r.Header.Get("tms-named-user"))
						body := &trackedBody{Reader: strings.NewReader("reply")}
						bodies = append(bodies, body)
						return &http.Response{StatusCode: status, Body: body}, nil
					}),
				}
				req, err := http.NewRequestWithContext(context.Background(), method,
					"https://api.test", strings.NewReader(`{"payload":true}`))
				require.NoError(t, err)
				req.Header.Set("tms-named-user", "automation")
				resp, err := transport.RoundTrip(req)
				require.NoError(t, err)
				expected := 1
				if status == 429 || status >= 500 {
					expected = 4
				}
				require.Equal(t, expected, calls)
				for _, body := range bodies[:len(bodies)-1] {
					require.True(t, body.closed)
				}
				require.False(t, bodies[len(bodies)-1].closed)
				require.NoError(t, resp.Body.Close())
			})
		}
	}
}

func TestRetryCancellationAndTransportErrors(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	calls := 0
	body := &trackedBody{Reader: strings.NewReader("")}
	transport := &retryTransport{
		maxRetries: 3,
		retryDelay: time.Hour,
		wrapped: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			calls++
			cancel()
			return &http.Response{StatusCode: 503, Body: body}, nil
		}),
	}
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.test", nil)
	require.NoError(t, err)
	_, err = transport.RoundTrip(req)
	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, 1, calls)
	require.True(t, body.closed)

	failure := errors.New("network failure")
	calls = 0
	transport.wrapped = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return nil, failure
	})
	req, err = http.NewRequestWithContext(context.Background(), "GET", "https://api.test", nil)
	require.NoError(t, err)
	_, err = transport.RoundTrip(req)
	require.ErrorIs(t, err, failure)
	require.Equal(t, 1, calls)
}

func TestTimeoutAndCancelledCalls(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		<-r.Context().Done()
		return nil, r.Context().Err()
	})
	cfg := CtmsClientConfig{Endpoint: "https://api.test", TokenURL: "https://auth.test/token"}
	c := NewCtmsClientFacadeWithHTTP(cfg, &http.Client{
		Timeout:   10 * time.Millisecond,
		Transport: transport,
	})
	_, err := c.Nodes.Get(context.Background(), 1)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = c.Nodes.Get(ctx, 1)
	require.ErrorIs(t, err, context.Canceled)
}

func TestTokenRetriesReplayCredentialsAndSkipResourceAttribution(t *testing.T) {
	var tokens int
	transport := &retryTransport{
		maxRetries: 3,
		wrapped: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.URL.Host != "auth.test" {
				return response(200, `{"id":1}`), nil
			}
			tokens++
			assert.Empty(t, r.Header.Get("tms-named-user"))
			id, secret, ok := r.BasicAuth()
			assert.True(t, ok)
			assert.Equal(t, "id", id)
			assert.Equal(t, "secret", secret)
			body, err := io.ReadAll(r.Body)
			assert.NoError(t, err)
			assert.NoError(t, r.Body.Close())
			assert.Equal(t, "grant_type=client_credentials", string(body))
			if tokens == 1 {
				return response(429, "try again"), nil
			}
			return response(200, `{"access_token":"token","expires_in":3600}`), nil
		}),
	}
	cfg := CtmsClientConfig{
		Endpoint:     "https://api.test",
		TokenURL:     "https://auth.test/token",
		ClientID:     "id",
		ClientSecret: "secret",
		NamedUser:    "automation",
	}
	c := NewCtmsClientFacadeWithHTTP(cfg, &http.Client{Transport: transport})
	_, err := c.Nodes.Get(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, 2, tokens)
}

func TestTokenWithoutLifetimeIsNotReused(t *testing.T) {
	var tokens int
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host == "auth.test" {
			tokens++
			return response(200, `{"access_token":"token"}`), nil
		}
		return response(200, `{"id":1}`), nil
	})
	cfg := CtmsClientConfig{Endpoint: "https://api.test", TokenURL: "https://auth.test/token"}
	c := NewCtmsClientFacadeWithHTTP(cfg, &http.Client{Transport: transport})
	for range 2 {
		_, err := c.Nodes.Get(context.Background(), 1)
		require.NoError(t, err)
	}
	require.Equal(t, 2, tokens)
}

func TestRetryStopsWithoutFinalWait(t *testing.T) {
	// A long configured delay must not delay an already exhausted request.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	transport := &retryTransport{
		maxRetries: 0,
		retryDelay: time.Hour,
		wrapped: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return response(503, "unavailable"), nil
		}),
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.test", nil)
	require.NoError(t, err)
	resp, err := transport.RoundTrip(req)
	require.NoError(t, err)
	require.Equal(t, 503, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}
