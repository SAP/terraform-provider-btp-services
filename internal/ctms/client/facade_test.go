package ctmsclient

import (
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Local token endpoint plus an injected HTTP client, following the CI/CD tests.
func prepareClientFacadeForTest(t *testing.T, namedUser string, handler http.HandlerFunc) *CtmsClientFacade {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Empty(t, r.Header.Get("tms-named-user"))
			id, secret, ok := r.BasicAuth()
			assert.True(t, ok)
			assert.Equal(t, "test-id", id)
			assert.Equal(t, "test-secret", secret)
			assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
			assert.NoError(t, r.ParseForm())
			assert.Equal(t, "client_credentials", r.Form.Get("grant_type"))
			_, _ = w.Write([]byte(`{"access_token":"test-token","expires_in":3600}`))
			return
		}
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "terraform/test", r.Header.Get("User-Agent"))
		handler(w, r)
	}))
	t.Cleanup(srv.Close)
	cfg := CtmsClientConfig{
		Endpoint:     srv.URL + "/",
		TokenURL:     srv.URL + "/oauth/token",
		ClientID:     "test-id",
		ClientSecret: "test-secret",
		NamedUser:    namedUser,
		UserAgent:    "terraform/test",
	}
	return NewCtmsClientFacadeWithHTTP(cfg, srv.Client())
}

// assertRequest verifies method and escaped path, including destination names.
func assertRequest(t *testing.T, r *http.Request, method, path string) {
	t.Helper()
	assert.Equal(t, method, r.Method)
	assert.Equal(t, path, r.URL.EscapedPath())
}

// assertRequestBody decodes a request body, following the CI/CD test helper.
func assertRequestBody(t *testing.T, r *http.Request, dst any) {
	t.Helper()
	assert.NoError(t, json.NewDecoder(r.Body).Decode(dst))
}

func assertEmptyRequestBody(t *testing.T, r *http.Request) {
	t.Helper()
	body, err := io.ReadAll(r.Body)
	assert.NoError(t, err)
	assert.Empty(t, body)
}

// Compare against literal API JSON so missing model fields cannot mask a failure.
func assertResponseBody(t *testing.T, response any, want string) {
	t.Helper()
	body, err := json.Marshal(response)
	require.NoError(t, err)
	require.JSONEq(t, want, string(body))
}

// prepareLandscapeFacadeForTest provides state for facade-specific lifecycle tests.
// It verifies the wire contract, not live API reset behavior or Terraform plans.
func prepareLandscapeFacadeForTest(t *testing.T, collection string) (*CtmsClientFacade, func(string)) {
	t.Helper()
	var mu sync.Mutex
	var stored map[string]json.RawMessage
	client := prepareClientFacadeForTest(t, "automation", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		assert.True(t, strings.HasPrefix(r.URL.Path, "/v2/"+collection))
		if r.Method != http.MethodPost && stored == nil {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"gone"}`))
			return
		}
		switch r.Method {
		case http.MethodPost:
			assert.NoError(t, json.NewDecoder(r.Body).Decode(&stored))
			stored["id"] = json.RawMessage("1")
			w.WriteHeader(http.StatusCreated)
		case http.MethodPatch:
			var patch map[string]json.RawMessage
			assert.NoError(t, json.NewDecoder(r.Body).Decode(&patch))
			maps.Copy(stored, patch)
		case http.MethodDelete:
			stored = nil
			w.WriteHeader(http.StatusNoContent)
			return
		}
		assert.NoError(t, json.NewEncoder(w).Encode(stored))
	})
	setDescription := func(description string) {
		mu.Lock()
		defer mu.Unlock()
		encoded, err := json.Marshal(description)
		require.NoError(t, err)
		stored["description"] = encoded
	}
	return client, setDescription
}

// A separate client demonstrates import/read without prior create history.
func newFacadeForImportTest(client *CtmsClientFacade) *CtmsClientFacade {
	return NewCtmsClientFacadeWithHTTP(CtmsClientConfig{
		Endpoint:     client.httpClient.baseURL,
		TokenURL:     client.httpClient.tokenURL,
		ClientID:     "test-id",
		ClientSecret: "test-secret",
		NamedUser:    "automation",
		UserAgent:    "terraform/test",
	}, client.httpClient.httpClient)
}
