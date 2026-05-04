// btpservices/provider/testutil/vcr.go

package testutil

import (
	"net/http"
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
// the real value when recording (e.g. "client_id" → "SAPBTP_CICD_CLIENT_ID").
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

	rec.SetMatcher(defaultRequestMatcher(t))
	rec.AddHook(hookRedactSensitiveData(), recorder.BeforeSaveHook)
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
// differ between recording and replay.
func defaultRequestMatcher(t *testing.T) func(r *http.Request, i cassette.Request) bool {
	t.Helper()
	return func(r *http.Request, i cassette.Request) bool {
		return r.Method == i.Method && r.URL.String() == i.URL
	}
}

// hookRedactSensitiveData strips credentials and tokens from cassette bodies
// before they are written to disk. The list of fields covers OAuth2 responses
// and any service that uses password-style credentials.
func hookRedactSensitiveData() func(i *cassette.Interaction) error {
	return func(i *cassette.Interaction) error {
		redactJSONField(&i.Response.Body, "access_token", "redacted-access-token")
		redactJSONField(&i.Response.Body, "refresh_token", "redacted-refresh-token")
		redactJSONField(&i.Response.Body, "client_id", "redacted-client-id")
		redactJSONField(&i.Response.Body, "client_secret", "redacted-client-secret")
		redactJSONField(&i.Request.Body, "password", "redacted-password")
		redactJSONField(&i.Response.Body, "password", "redacted-password")
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
