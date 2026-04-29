// internal/services/cicd/models/credential.go

package cicdmodels

// Credential is the API response model for a single credential.
// Sensitive fields (password, token, etc.) are never returned by the API on read.
// The type-specific readable fields are surfaced via the nested model fields below.
type Credential struct {
	ID          string `json:"id"`
	APIVersion  string `json:"apiVersion"`
	Name        string `json:"name"`
	Description string `json:"description"`

	// BasicAuth is populated only when the credential is of BasicAuth type.
	// On reads the API returns the username but NOT the password.
	Basic *BasicAuthModel `json:"basic,omitempty"`
}

// BasicAuthModel is the read-response sub-object for basic-auth credentials.
// The password field is intentionally absent — the API never returns it.
type BasicAuthModel struct {
	Username string `json:"username"`
}

// CreateCredentialRequest is the body sent to POST /v2/credentials.
// Exactly one of the typed sub-objects must be set.
type CreateCredentialRequest struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Basic       *BasicAuth `json:"basic,omitempty"`
}

// UpdateCredentialRequest is the body sent to PUT /v2/credentials/{reference}.
// The name field is immutable; only description and the typed sub-object can change.
type UpdateCredentialRequest struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Basic       *BasicAuth `json:"basic,omitempty"`
}

// PatchCredentialRequest is the body sent to PATCH /v2/credentials/{reference}.
// Only non-nil fields are sent — omitempty ensures zero values are omitted.
// Use pointer fields so callers can distinguish "set to empty" from "not provided".
type PatchCredentialRequest struct {
	Name        *string    `json:"name,omitempty"`
	Description *string    `json:"description,omitempty"`
	Basic       *BasicAuth `json:"basic,omitempty"`
}

// BasicAuth is the write payload for the basic-auth credential sub-type.
type BasicAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// CredentialListResponse is the envelope returned by GET /v2/credentials.
type CredentialListResponse struct {
	Embedded *CredentialListEmbedded `json:"_embedded,omitempty"`
}

type CredentialListEmbedded struct {
	Credentials []Credential `json:"credentials"`
}
