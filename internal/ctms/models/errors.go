package ctmsmodels

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// APIError preserves the HTTP status and CTMS details for provider diagnostics.
// This is handwritten support, separate from the generated wire models.
type APIError struct {
	StatusCode int
	Reference  string
	ErrorType  string
	Message    string
}

func (e *APIError) Error() string {
	message := fmt.Sprintf("CTMS API error %d for %s", e.StatusCode, e.Reference)
	if e.ErrorType != "" {
		message += ": " + e.ErrorType
	}
	if e.Message != "" {
		message += ": " + e.Message
	}
	return message
}

// NotFoundError signals a 404. Resources decide whether to remove state or
// accept an already deleted object; data sources can report a missing object.
type NotFoundError struct {
	*APIError
}

func (e *NotFoundError) Unwrap() error {
	return e.APIError
}

// IsNotFound reports whether err wraps a CTMS 404 response.
func IsNotFound(err error) bool {
	var target *NotFoundError
	return errors.As(err, &target)
}

// CheckAPIResponse returns a typed error for non-success responses.
// It consumes the error body; the caller remains responsible for closing it.
func CheckAPIResponse(resp *http.Response, reference string) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	body, readErr := io.ReadAll(resp.Body)
	var parsed struct {
		TmsExceptionObject
		// NotAuthorizedException is not present in the generated models.
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	_ = json.Unmarshal(body, &parsed)
	err := &APIError{
		StatusCode: resp.StatusCode,
		Reference:  reference,
		ErrorType:  parsed.ErrorType,
		Message:    parsed.Message,
	}
	if err.ErrorType == "" {
		err.ErrorType = parsed.Error
	}
	if err.Message == "" {
		err.Message = parsed.ErrorDescription
	}
	if err.Message == "" {
		err.Message = strings.TrimSpace(string(body))
	}
	if readErr != nil {
		err.Message += fmt.Sprintf(" (read error body: %v)", readErr)
	}
	if resp.StatusCode == http.StatusNotFound {
		return &NotFoundError{APIError: err}
	}
	return err
}
