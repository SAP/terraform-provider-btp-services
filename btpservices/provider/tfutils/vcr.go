// btpservices/provider/tfutils/vcr.go

package tfutils

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"slices"
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
//
// Credential body redaction uses a structural deny-all approach: any value nested
// under a known credential wrapper key is replaced, and any field whose name matches
// a secret-name pattern is also replaced — regardless of JSON nesting depth.
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

		// Credential bodies: structural deny-all redaction.
		// Redact all string values nested under known credential wrapper objects,
		// and any field whose name matches a secret pattern at any depth.
		redactCredentialBodies(&i.Request.Body)
		redactCredentialBodies(&i.Response.Body)

		// API response body: redact _links
		redactJSONLinks(&i.Response.Body)

		// Response headers
		redactResponseHeader(i.Response.Headers, "X-Vcap-Request-Id", "redacted-vcap-request-id")
		// Location header: replace only the hostname so the path (including resource ID) is preserved.
		redactLocationHeader(i.Response.Headers, apiHost, "redacted-api-host")

		return nil
	}
}

// credentialWrapperKeys are top-level JSON keys whose entire nested object contains
// credential secrets. All string leaf values beneath them are redacted.
var credentialWrapperKeys = []string{
	"basic",
	"webhookToken",
	"serviceKey",
	"secretText",
	"kubernetesConfiguration",
	"containerRegistryConfiguration",
}

// secretFieldPatterns are field name substrings that indicate a secret value
// at any JSON nesting depth.
var secretFieldPatterns = []string{
	"password", "token", "secret", "key", "credential", "kubeconfig",
}

// redactCredentialBodies applies structural deny-all redaction to a JSON body string.
// It unmarshals the JSON, walks the object tree, and redacts:
//  1. All string values inside known credential wrapper objects.
//  2. All string values whose field name matches a secret pattern.
//
// If the body is not valid JSON it is left unchanged.
func redactCredentialBodies(body *string) {
	if body == nil || *body == "" {
		return
	}
	var obj any
	if err := json.Unmarshal([]byte(*body), &obj); err != nil {
		return // not JSON — leave as-is
	}
	redactJSONValue(&obj, false)
	b, err := json.Marshal(obj)
	if err != nil {
		return
	}
	*body = string(b)
}

// redactJSONValue walks a decoded JSON value and redacts sensitive string fields.
// inWrapper is true when we are already inside a known credential wrapper object —
// in that case every string leaf value is replaced.
func redactJSONValue(v *any, inWrapper bool) {
	switch val := (*v).(type) {
	case map[string]any:
		for k, child := range val {
			childInWrapper := inWrapper || isCredentialWrapper(k)
			if childInWrapper {
				// Replace every string leaf inside the wrapper.
				redactJSONValue(&child, true)
			} else if isSecretField(k) {
				// Replace this field's string value regardless of depth.
				if _, ok := child.(string); ok {
					child = "redacted"
				} else {
					redactJSONValue(&child, false)
				}
			} else {
				redactJSONValue(&child, false)
			}
			val[k] = child
		}
	case []any:
		for i := range val {
			redactJSONValue(&val[i], inWrapper)
		}
	case string:
		if inWrapper {
			*v = "redacted"
		}
	}
}

func isCredentialWrapper(key string) bool {
	return slices.Contains(credentialWrapperKeys, key)
}

func isSecretField(key string) bool {
	lower := strings.ToLower(key)
	for _, pat := range secretFieldPatterns {
		if strings.Contains(lower, pat) {
			return true
		}
	}
	return false
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

// redactLocationHeader replaces the hostname in a Location response header with
// a redacted placeholder, preserving the path so resource IDs remain intact.
func redactLocationHeader(headers map[string][]string, realHost, redactedHost string) {
	if realHost == "" {
		return
	}
	for key, vals := range headers {
		if strings.EqualFold(key, "location") {
			for i, v := range vals {
				vals[i] = strings.ReplaceAll(v, realHost, redactedHost)
			}
			headers[key] = vals
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
