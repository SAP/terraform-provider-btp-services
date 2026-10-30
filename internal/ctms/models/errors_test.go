package ctmsmodels

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckAPIResponse(t *testing.T) {
	for _, tc := range []struct {
		name                string
		status              int
		body, kind, message string
	}{
		{"success", 201, `{"id":1}`, "", ""},
		{"no content", 204, "", "", ""},
		{"missing", 404, `{"errorType":"NotFound","message":"node gone"}`, "NotFound", "node gone"},
		{"unauthorized", 401, `{"error":"unauthorized","error_description":"token required"}`, "unauthorized", "token required"},
		{"forbidden", 403, `{"error":"forbidden","error_description":"missing permission"}`, "forbidden", "missing permission"},
		{"unprocessable", 422, `{"errorType":"NodeInUse","message":"node still in route"}`, "NodeInUse", "node still in route"},
		{"text", 500, "backend unavailable", "", "backend unavailable"},
		{"empty", 404, "", "", ""},
		{"malformed", 502, `{broken`, "", `{broken`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp := &http.Response{StatusCode: tc.status, Body: io.NopCloser(strings.NewReader(tc.body))}
			defer func() { _ = resp.Body.Close() }()
			err := CheckAPIResponse(resp, "/v2/nodes/1")
			if tc.status < 300 {
				require.NoError(t, err)
				body, readErr := io.ReadAll(resp.Body)
				require.NoError(t, readErr)
				require.Equal(t, tc.body, string(body))
				return
			}
			require.Error(t, err)
			wrapped := fmt.Errorf("read resource: %w", err)
			require.Equal(t, tc.status == 404, IsNotFound(wrapped))
			var api *APIError
			require.True(t, errors.As(wrapped, &api))
			require.Equal(t, tc.status, api.StatusCode)
			require.Equal(t, tc.kind, api.ErrorType)
			require.Equal(t, tc.message, api.Message)
			require.Contains(t, api.Error(), "/v2/nodes/1")
			require.Contains(t, api.Error(), fmt.Sprint(tc.status))
		})
	}
	require.False(t, IsNotFound(nil))
}
