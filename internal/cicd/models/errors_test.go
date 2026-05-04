// internal/cicd/models/errors_test.go

package cicdmodels

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestCheckAPIResponse_2xx_ReturnsNil(t *testing.T) {
	for _, code := range []int{200, 201, 204} {
		err := CheckAPIResponse(makeResponse(code, ""), "ref")
		assert.NoError(t, err, "expected no error for status %d", code)
	}
}

func TestCheckAPIResponse_404_NotFoundError(t *testing.T) {
	body := `{
		"type": "about:blank",
		"title": "A requested entity was not found.",
		"status": 404,
		"detail": "No credential with name or ID 'my-github-token-update' found.",
		"instance": "/v2/credentials/my-github-token-update",
		"correlationID": "87e0bdb8-086a-4b14-6f7d-0921bfe8b175"
	}`

	err := CheckAPIResponse(makeResponse(http.StatusNotFound, body), "my-github-token-update")

	require.Error(t, err)
	assert.True(t, IsNotFound(err))

	var nfe *NotFoundError
	require.ErrorAs(t, err, &nfe)
	assert.Equal(t, "my-github-token-update", nfe.Reference)
	assert.Equal(t, "A requested entity was not found.", nfe.Title)
	assert.Contains(t, nfe.Error(), "No credential with name or ID 'my-github-token-update' found.")
}

func TestCheckAPIResponse_400_ValidationFieldErrors(t *testing.T) {
	body := `{
		"type": "about:blank",
		"title": "Validation failed for received resource 'RepositoryModel'.",
		"status": 400,
		"instance": "/v2/repositories",
		"correlationID": "3664016d-f761-4b97-518b-85be14f3c963",
		"fieldErrors": [
			{"fieldName": "name", "reason": "must match \"[a-zA-Z0-9_-]{1,64}\"."}
		]
	}`

	err := CheckAPIResponse(makeResponse(http.StatusBadRequest, body), "/v2/repositories")

	require.Error(t, err)
	assert.False(t, IsNotFound(err))

	var apiErr *cicdAPIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 400, apiErr.StatusCode)
	assert.Equal(t, "Validation failed for received resource 'RepositoryModel'.", apiErr.Title)
	assert.Equal(t, "3664016d-f761-4b97-518b-85be14f3c963", apiErr.CorrelationID)
	require.Len(t, apiErr.FieldErrors, 1)
	assert.Equal(t, "name", apiErr.FieldErrors[0].FieldName)
	assert.Contains(t, apiErr.FieldErrors[0].Reason, "[a-zA-Z0-9_-]{1,64}")

	msg := err.Error()
	assert.Contains(t, msg, "400")
	assert.Contains(t, msg, "Validation failed")
	assert.Contains(t, msg, `field "name"`)
	assert.Contains(t, msg, "[a-zA-Z0-9_-]{1,64}")
	assert.Contains(t, msg, "3664016d-f761-4b97-518b-85be14f3c963")
}

func TestCheckAPIResponse_500_InternalServerError(t *testing.T) {
	body := `{
		"type": "about:blank",
		"title": "Internal Server Error",
		"status": 500,
		"detail": "We experienced an unexpected issue. Please raise a ticket and provide the correlation id.",
		"instance": "/v2/repositories",
		"correlationID": "b0c2ec45-a71e-4d59-6846-666375a57cc3"
	}`

	err := CheckAPIResponse(makeResponse(http.StatusInternalServerError, body), "/v2/repositories")

	require.Error(t, err)
	assert.False(t, IsNotFound(err))

	var apiErr *cicdAPIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 500, apiErr.StatusCode)
	assert.Equal(t, "Internal Server Error", apiErr.Title)
	assert.Contains(t, apiErr.Detail, "unexpected issue")
	assert.Equal(t, "b0c2ec45-a71e-4d59-6846-666375a57cc3", apiErr.CorrelationID)

	msg := err.Error()
	assert.Contains(t, msg, "500")
	assert.Contains(t, msg, "Internal Server Error")
	assert.Contains(t, msg, "unexpected issue")
	assert.Contains(t, msg, "b0c2ec45-a71e-4d59-6846-666375a57cc3")
}

func TestCheckAPIResponse_NonJSON_StillReturnsError(t *testing.T) {
	err := CheckAPIResponse(makeResponse(http.StatusBadGateway, "<html>Bad Gateway</html>"), "/v2/credentials")

	require.Error(t, err)
	assert.False(t, IsNotFound(err))

	var apiErr *cicdAPIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 502, apiErr.StatusCode)
	// title/detail are empty — JSON parse failed silently
	assert.Empty(t, apiErr.Title)
	assert.Contains(t, apiErr.Error(), "502")
}
