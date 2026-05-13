// btpservices/provider/tfutils/vcr.go

package tfutils

import (
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

// TestCredentials is a generic key→value map of service config values.
// Each service defines its own keys and redacted placeholder values.
// Example keys for CI/CD: "endpoint", "token_url", "client_id", "client_secret".
type TestCredentials map[string]string

// SetupVCR creates a VCR recorder for acceptance tests.
//
// liveEnvVars maps each credential key to the environment variable that holds
// the real value when recording (e.g. "client_id" → "BTP_CICD_CLIENT_ID").
// redacted holds the safe placeholder values that are written into cassettes
// and used on replay — no live credentials are needed after the first recording.
//
// Set TEST_FORCE_REC=true to force re-recording even when a cassette exists.
func SetupVCR(t *testing.T, cassetteName string, liveEnvVars map[string]string, redacted TestCredentials) (*recorder.Recorder, TestCredentials) {
	t.Helper()

	mode := recorder.ModeRecordOnce
	if force, _ := strconv.ParseBool(os.Getenv("TEST_FORCE_REC")); force {
		mode = recorder.ModeRecordOnly
	}

	rec, err := recorder.NewWithOptions(&recorder.Options{
		CassetteName:       cassetteName,
		Mode:               mode,
		SkipRequestLatency: true,
		RealTransport:      http.DefaultTransport,
	})
	if err != nil {
		t.Fatalf("failed to create VCR recorder: %v", err)
	}

	creds := redacted

	if rec.IsRecording() {
		t.Logf("ATTENTION: Recording cassette '%s'", cassetteName)

		live := make(TestCredentials, len(liveEnvVars))
		for key, envVar := range liveEnvVars {
			val := os.Getenv(envVar)
			if val == "" {
				t.Fatalf("env var %s (required for key %q) is not set — cannot record cassette", envVar, key)
			}
			live[key] = val
		}
		creds = live
	} else {
		t.Logf("Replaying cassette '%s'", cassetteName)
	}

	rec.SetMatcher(defaultRequestMatcher(t, creds))
	rec.AddHook(hookRedactSensitiveData(creds), recorder.BeforeSaveHook)
	rec.AddHook(hookRedactAuthHeaders(), recorder.BeforeSaveHook)

	return rec, creds
}

// StopQuietly stops the recorder, panicking only on unexpected errors.
func StopQuietly(rec *recorder.Recorder) {
	if err := rec.Stop(); err != nil {
		panic(err)
	}
}

// defaultRequestMatcher matches recorded interactions on HTTP method + URL.
// Authorization headers are intentionally excluded — they contain tokens that
// differ between recording and replay. Real hostnames are normalised to their
// redacted placeholders before comparison so replay works after redaction.
func defaultRequestMatcher(t *testing.T, creds TestCredentials) func(r *http.Request, i cassette.Request) bool {
	t.Helper()
	tokenHost := hostOf(creds["token_url"])
	apiHost := hostOf(creds["endpoint"])
	normalise := func(u string) string {
		if tokenHost != "" {
			u = strings.ReplaceAll(u, tokenHost, "redacted-token-host")
		}
		if apiHost != "" {
			u = strings.ReplaceAll(u, apiHost, "redacted-api-host")
		}
		return u
	}
	return func(r *http.Request, i cassette.Request) bool {
		return r.Method == i.Method && normalise(r.URL.String()) == i.URL
	}
}

// hookRedactSensitiveData strips credentials, tokens, and environment-specific
// values from cassette bodies and request metadata before they are written to disk.
// creds contains the live values so that hostnames can be replaced regardless of
// which environment was used during recording.
func hookRedactSensitiveData(creds TestCredentials) func(i *cassette.Interaction) error {
	// Extract hostnames from live endpoint and token_url so they are redacted
	// even if they change between environments.
	tokenHost := hostOf(creds["token_url"])
	apiHost := hostOf(creds["endpoint"])

	return func(i *cassette.Interaction) error {
		// Request host and url fields
		if tokenHost != "" {
			if i.Request.Host == tokenHost {
				i.Request.Host = "redacted-token-host"
			}
			i.Request.URL = strings.ReplaceAll(i.Request.URL, tokenHost, "redacted-token-host")
		}
		if apiHost != "" {
			if i.Request.Host == apiHost {
				i.Request.Host = "redacted-api-host"
			}
			i.Request.URL = strings.ReplaceAll(i.Request.URL, apiHost, "redacted-api-host")
		}

		// OAuth2 token response body
		redactJSONField(&i.Response.Body, "access_token", "redacted-access-token")
		redactJSONField(&i.Response.Body, "refresh_token", "redacted-refresh-token")
		redactJSONField(&i.Response.Body, "client_id", "redacted-client-id")
		redactJSONField(&i.Response.Body, "client_secret", "redacted-client-secret")
		redactJSONField(&i.Response.Body, "scope", "redacted-scope")
		redactJSONField(&i.Response.Body, "jti", "redacted-jti")

		// Credential request/response body
		redactJSONField(&i.Request.Body, "password", "redacted-password")
		redactJSONField(&i.Response.Body, "password", "redacted-password")

		// API response body: redact _links
		redactJSONLinks(&i.Response.Body)

		// Response headers
		redactResponseHeader(i.Response.Headers, "X-Vcap-Request-Id", "redacted-vcap-request-id")
		redactResponseHeader(i.Response.Headers, "Location", "redacted-location")

		return nil
	}
}

// hookRedactAuthHeaders removes Authorization and session headers from saved cassettes.
func hookRedactAuthHeaders() func(i *cassette.Interaction) error {
	return func(i *cassette.Interaction) error {
		redactHeaders(i.Request.Headers)
		redactHeaders(i.Response.Headers)
		return nil
	}
}

func redactHeaders(headers map[string][]string) {
	for key := range headers {
		lower := strings.ToLower(key)
		if lower == "authorization" ||
			strings.Contains(lower, "token") ||
			strings.Contains(lower, "session") {
			headers[key] = []string{"redacted"}
		}
	}
}

func redactResponseHeader(headers map[string][]string, name, replacement string) {
	for key := range headers {
		if strings.EqualFold(key, name) {
			headers[key] = []string{replacement}
		}
	}
}

// hostOf returns the hostname from a URL string, or empty string on error.
func hostOf(rawURL string) string {
	if rawURL == "" {
		return ""
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return u.Hostname()
}

// redactJSONLinks replaces the entire _links object value with a redacted placeholder.
func redactJSONLinks(body *string) {
	if body == nil {
		return
	}
	const key = `"_links":`
	const replacement = `"_links":{"self":{"href":"redacted"}}`
	result := *body
	searchFrom := 0
	for {
		start := strings.Index(result[searchFrom:], key)
		if start < 0 {
			break
		}
		start += searchFrom
		objStart := start + len(key)
		if objStart >= len(result) || result[objStart] != '{' {
			searchFrom = start + len(key)
			continue
		}
		depth := 0
		end := -1
		for i := objStart; i < len(result); i++ {
			switch result[i] {
			case '{':
				depth++
			case '}':
				depth--
				if depth == 0 {
					end = i + 1
				}
			}
			if end > 0 {
				break
			}
		}
		if end < 0 {
			break
		}
		result = result[:start] + replacement + result[end:]
		// Advance past the replacement to avoid re-matching it.
		searchFrom = start + len(replacement)
	}
	*body = result
}

// redactJSONField replaces the string value of a single JSON field in-place.
func redactJSONField(body *string, field, replacement string) {
	if body == nil {
		return
	}
	needle := `"` + field + `":"`
	start := strings.Index(*body, needle)
	if start < 0 {
		return
	}
	valueStart := start + len(needle)
	valueEnd := strings.Index((*body)[valueStart:], `"`)
	if valueEnd < 0 {
		return
	}
	*body = (*body)[:valueStart] + replacement + (*body)[valueStart+valueEnd:]
}
