// internal/cicd/models/errors.go

package cicdmodels

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// NotFoundError signals a 404 response from the CI/CD API.
type NotFoundError struct {
	Reference string
	Title     string
	Detail    string
}

func (e *NotFoundError) Error() string {
	if e.Detail != "" {
		return e.Detail
	}
	return fmt.Sprintf("resource %q not found", e.Reference)
}

// IsNotFound returns true if err wraps a NotFoundError.
func IsNotFound(err error) bool {
	var nfe *NotFoundError
	return errors.As(err, &nfe)
}

// apiErrorBody is the structured error body returned by the CI/CD API.
type apiErrorBody struct {
	Title         string       `json:"title"`
	Status        int          `json:"status"`
	Detail        string       `json:"detail"`
	CorrelationID string       `json:"correlationID"`
	FieldErrors   []fieldError `json:"fieldErrors"`
}

type fieldError struct {
	FieldName string `json:"fieldName"`
	Reason    string `json:"reason"`
}

type cicdAPIError struct {
	StatusCode    int
	Title         string
	Detail        string
	CorrelationID string
	FieldErrors   []fieldError
}

func (e *cicdAPIError) Error() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "CI/CD API error %d", e.StatusCode)
	if e.Title != "" {
		fmt.Fprintf(&sb, ": %s", e.Title)
	}
	if e.Detail != "" {
		fmt.Fprintf(&sb, "\n  detail: %s", e.Detail)
	}
	for _, fe := range e.FieldErrors {
		fmt.Fprintf(&sb, "\n  - field %q: %s", fe.FieldName, fe.Reason)
	}
	if e.CorrelationID != "" {
		fmt.Fprintf(&sb, "\n  correlation ID: %s", e.CorrelationID)
	}
	return sb.String()
}

// CheckAPIResponse inspects an HTTP response and returns a typed error when
// the status code indicates failure. Call this after every HTTP response.
func CheckAPIResponse(resp *http.Response, reference string) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)

	var parsed apiErrorBody
	_ = json.Unmarshal(body, &parsed)

	if resp.StatusCode == http.StatusNotFound {
		return &NotFoundError{
			Reference: reference,
			Title:     parsed.Title,
			Detail:    parsed.Detail,
		}
	}

	return &cicdAPIError{
		StatusCode:    resp.StatusCode,
		Title:         parsed.Title,
		Detail:        parsed.Detail,
		CorrelationID: parsed.CorrelationID,
		FieldErrors:   parsed.FieldErrors,
	}
}
