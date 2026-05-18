// internal/cicd/models/repository.go

package cicdmodels

// Repository is the API response model for a single repository.
type Repository struct {
	ID                         string              `json:"id"`
	APIVersion                 string              `json:"apiVersion"`
	Name                       string              `json:"name"`
	CloneURL                   string              `json:"cloneUrl"`
	CloneCredentialID          *string             `json:"cloneCredentialId,omitempty"`
	CloudConnectorCredentialID *string             `json:"cloudConnectorCredentialId,omitempty"`
	EventReceiver              *EventReceiverModel `json:"eventReceiver,omitempty"`
}

// EventReceiverModel is the event receiver configuration embedded in a repository.
// webhookId is immutable and assigned by the API on creation.
type EventReceiverModel struct {
	Active                   bool   `json:"active"`
	SCMType                  string `json:"scmType"`
	WebhookID                string `json:"webhookId,omitempty"`
	WebhookTokenCredentialID string `json:"webhookTokenCredentialId,omitempty"`
}

// CreateRepositoryRequest is the body sent to POST /v2/repositories.
// Required fields: name, cloneUrl.
type CreateRepositoryRequest struct {
	Name                       string              `json:"name"`
	CloneURL                   string              `json:"cloneUrl"`
	CloneCredentialID          *string             `json:"cloneCredentialId,omitempty"`
	CloudConnectorCredentialID *string             `json:"cloudConnectorCredentialId,omitempty"`
	EventReceiver              *EventReceiverModel `json:"eventReceiver,omitempty"`
}

// UpdateRepositoryRequest is the body sent to PUT /v2/repositories.
// The full model including id is required — the API uses id in the body to identify the repository.
// Returns 204 with no body.
type UpdateRepositoryRequest struct {
	ID                         string              `json:"id"`
	Name                       string              `json:"name"`
	CloneURL                   string              `json:"cloneUrl"`
	CloneCredentialID          *string             `json:"cloneCredentialId,omitempty"`
	CloudConnectorCredentialID *string             `json:"cloudConnectorCredentialId,omitempty"`
	EventReceiver              *EventReceiverModel `json:"eventReceiver,omitempty"`
}

// RepositoryListResponse is the envelope returned by GET /v2/repositories.
type RepositoryListResponse struct {
	Embedded *RepositoryListEmbedded `json:"_embedded,omitempty"`
}

// RepositoryListEmbedded holds the nested repositories array.
type RepositoryListEmbedded struct {
	Repositories []Repository `json:"repositories"`
}
