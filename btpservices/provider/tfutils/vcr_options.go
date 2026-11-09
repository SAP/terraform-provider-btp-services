package tfutils

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

// VCROptions customizes recording without changing existing SetupVCR callers.
type VCROptions struct {
	// RequireExplicitRecording skips missing cassettes unless TEST_FORCE_REC enables recording.
	RequireExplicitRecording bool
	MatchRequestBody         bool
	// MatchHeaders must contain non-secret headers; custom redaction retains them in cassettes.
	MatchHeaders []string
	// Redaction replaces the default CI/CD redaction policy when set.
	Redaction *VCRRedaction
}

// VCRRedaction describes service-specific normalization applied before saving
// and when matching requests. Headers other than Content-Type and MatchHeaders
// are removed, along with parsed forms and connection metadata.
type VCRRedaction struct {
	// URLKeys identifies URL credentials whose full values and hosts should be replaced.
	URLKeys []string
	// NormalizeURL optionally replaces service-specific paths and query parameters.
	NormalizeURL func(*url.URL, TestCredentials)
	// RedactString handles JSON string fields, including service credentials and tokens.
	RedactString func(key, value string, creds TestCredentials) string
}

func configureVCR(t *testing.T, rec *recorder.Recorder, creds, redacted TestCredentials, options VCROptions) {
	t.Helper()

	matcher := defaultRequestMatcher(t, creds)
	normalizeBody := func(body string) string {
		redactCredentialBodies(&body)
		return body
	}

	if options.Redaction == nil {
		rec.AddHook(hookRedactSensitiveData(creds), recorder.BeforeSaveHook)
		rec.AddHook(hookRedactAuthHeaders(), recorder.BeforeSaveHook)
	} else {
		normalizeURL, body := options.Redaction.normalizers(t, creds, redacted)
		normalizeBody = body
		matcher = func(req *http.Request, saved cassette.Request) bool {
			return req.Method == saved.Method && normalizeURL(req.URL.String()) == saved.URL
		}
		rec.AddHook(normalizedInteractionHook(normalizeURL, normalizeBody, options.MatchHeaders), recorder.BeforeSaveHook)
	}

	rec.SetMatcher(func(req *http.Request, saved cassette.Request) bool {
		if !matcher(req, saved) {
			return false
		}

		for _, header := range options.MatchHeaders {
			if req.Header.Get(header) != saved.Headers.Get(header) {
				return false
			}
		}

		if !options.MatchRequestBody {
			return true
		}

		var body []byte
		if req.Body != nil {
			var err error
			body, err = io.ReadAll(req.Body)
			req.Body = io.NopCloser(bytes.NewReader(body))
			if err != nil {
				return false
			}
		}

		return normalizeBody(string(body)) == saved.Body
	})
}

func (r *VCRRedaction) normalizers(t *testing.T, creds, redacted TestCredentials) (func(string) string, func(string) string) {
	t.Helper()

	var pairs []string
	for _, key := range r.URLKeys {
		actual := strings.TrimRight(creds[key], "/")
		placeholder := strings.TrimRight(redacted[key], "/")
		if actual == "" || placeholder == "" {
			t.Fatalf("URL redaction requires actual and placeholder values for %q", key)
		}

		u, err := url.Parse(actual)
		if err != nil {
			t.Fatal(err)
		}

		p, err := url.Parse(placeholder)
		if err != nil {
			t.Fatal(err)
		}

		pairs = append(pairs, actual, placeholder)
		if u.Host != "" {
			pairs = append(pairs, u.Host, p.Host)
		}
	}

	replacements := strings.NewReplacer(pairs...)
	normalizeURL := func(raw string) string {
		normalized := replacements.Replace(raw)
		parsed, err := url.Parse(normalized)
		if err != nil {
			return normalized
		}

		if r.NormalizeURL != nil {
			r.NormalizeURL(parsed, creds)
		}

		parsed.RawQuery = parsed.Query().Encode()
		return parsed.String()
	}

	normalizeBody := func(body string) string {
		return transformJSONBody(replacements.Replace(body), func(value *any) {
			if r.RedactString != nil {
				redactJSONStrings(value, "", func(key, value string) string {
					return r.RedactString(key, value, creds)
				})
			}
		})
	}

	return normalizeURL, normalizeBody
}

func redactJSONStrings(value *any, key string, redact func(string, string) string) {
	switch v := (*value).(type) {
	case map[string]any:
		for field, child := range v {
			redactJSONStrings(&child, field, redact)
			v[field] = child
		}
	case []any:
		for i := range v {
			redactJSONStrings(&v[i], key, redact)
		}
	case string:
		*value = redact(key, v)
	}
}

func normalizedInteractionHook(normalizeURL, normalizeBody func(string) string, matchHeaders []string) func(*cassette.Interaction) error {
	return func(i *cassette.Interaction) error {
		i.Request.URL = normalizeURL(i.Request.URL)
		i.Request.Host = normalizeURL(i.Request.Host)
		i.Request.Body = normalizeBody(i.Request.Body)
		i.Response.Body = normalizeBody(i.Response.Body)

		// Do not persist parsed OAuth forms or authentication/session metadata.
		i.Request.Form = nil
		i.Request.RequestURI = ""
		i.Request.RemoteAddr = ""
		i.Request.Trailer = nil
		i.Response.Trailer = nil
		keepHeaders(i.Request.Headers, append([]string{"Content-Type"}, matchHeaders...))
		keepHeaders(i.Response.Headers, []string{"Content-Type"})

		return nil
	}
}

func keepHeaders(headers http.Header, allowed []string) {
	for key := range headers {
		keep := false
		for _, name := range allowed {
			if strings.EqualFold(key, name) {
				keep = true
				break
			}
		}

		if !keep {
			delete(headers, key)
		}
	}
}
