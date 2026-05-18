package cicdmodels

// CredentialUsageUser holds the id/name/type of the job or repository that uses the credential.
type CredentialUsageUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"` // "job" or "repository"
}

// CredentialUsage represents a single entry in the usages list for a credential.
// The API wraps the identity fields under a nested "user" object.
type CredentialUsage struct {
	User CredentialUsageUser `json:"user"`
}

// CredentialUsageListResponse is the HAL envelope returned by
// GET /v2/credentials/{reference}/usages.
type CredentialUsageListResponse struct {
	Embedded *CredentialUsageEmbedded `json:"_embedded,omitempty"`
}

// CredentialUsageEmbedded holds the usages array within the HAL envelope.
type CredentialUsageEmbedded struct {
	Usages []CredentialUsage `json:"usages"`
}

// JobCredentialListResponse is the response returned by
// GET /v2/jobs/{jobReference}/credentials.
type JobCredentialListResponse struct {
	IDs []string `json:"ids"`
}
