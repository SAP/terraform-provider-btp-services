// internal/cicd/models/errors.go

package cicdmodels

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

// NotFoundError signals a 404 response from the CI/CD API.
type NotFoundError struct{ Reference string }

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("resource %q not found", e.Reference)
}

// IsNotFound returns true if err wraps a NotFoundError.
func IsNotFound(err error) bool {
	var nfe *NotFoundError
	return errors.As(err, &nfe)
}

type cicdAPIError struct {
	StatusCode int
	Body       string
}

func (e *cicdAPIError) Error() string {
	return fmt.Sprintf("CI/CD API error %d: %s", e.StatusCode, e.Body)
}

// CheckAPIResponse inspects an HTTP response and returns a typed error when
// the status code indicates failure. Call this after every HTTP response.
func CheckAPIResponse(resp *http.Response, reference string) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusNotFound {
		return &NotFoundError{Reference: reference}
	}
	return &cicdAPIError{StatusCode: resp.StatusCode, Body: string(body)}
}
