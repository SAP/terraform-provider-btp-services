// internal/cicd/client/facade_test.go

package cicdclient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// prepareClientFacadeForTest spins up an httptest.Server with a token endpoint
// and wires a CicdClientFacade to it. Retries are disabled so error-path tests
// complete instantly. Returns the facade and the server — callers do defer srv.Close().
func prepareClientFacadeForTest(t *testing.T, handleFn http.HandlerFunc) (*CicdClientFacade, *httptest.Server) {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"test-token","expires_in":3600}`))
			return
		}
		handleFn.ServeHTTP(w, r)
	}))

	cfg := CicdClientConfig{
		Endpoint:     srv.URL,
		TokenURL:     srv.URL + "/oauth/token",
		ClientID:     "test-id",
		ClientSecret: "test-secret",
	}
	facade := NewCicdClientFacadeWithHTTP(cfg, srv.Client())
	// Disable retry transport so error-path tests complete instantly.
	facade.httpClient.httpClient.Transport = http.DefaultTransport

	return facade, srv
}

// assertRequest verifies the HTTP method, path, and Content-Type of a request.
func assertRequest(t *testing.T, r *http.Request, expectedMethod, expectedPath string) {
	t.Helper()
	assert.Equal(t, expectedMethod, r.Method)
	assert.Equal(t, expectedPath, r.URL.Path)
}

// assertRequestBody decodes the JSON request body into dst and asserts no error.
func assertRequestBody(t *testing.T, r *http.Request, dst any) {
	t.Helper()
	err := json.NewDecoder(r.Body).Decode(dst)
	assert.NoError(t, err)
}
