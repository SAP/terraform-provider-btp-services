// internal/cicd/models/credential.go

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

	// CloudConnector is populated only when the credential is of Cloud Connector type.
	// locationId is readable on GET.
	CloudConnector *CloudConnectorModel `json:"cloudConnector,omitempty"`
}

// BasicAuthModel is the read-response sub-object for basic-auth credentials.
// The password field is intentionally absent — the API never returns it.
type BasicAuthModel struct {
	Username string `json:"username"`
}

// CreateCredentialRequest is the body sent to POST /v2/credentials.
// Exactly one of the typed sub-objects must be set.
type CreateCredentialRequest struct {
	Name                           string                          `json:"name"`
	Description                    string                          `json:"description"`
	Basic                          *BasicAuth                      `json:"basic,omitempty"`
	CloudConnector                 *CloudConnector                 `json:"cloudConnector,omitempty"`
	WebhookToken                   *WebhookToken                   `json:"webhookToken,omitempty"`
	ContainerRegistryConfiguration *ContainerRegistryConfiguration `json:"containerRegistryConfiguration,omitempty"`
	KubernetesConfiguration        *KubernetesConfiguration        `json:"kubernetesConfiguration,omitempty"`
}

// UpdateCredentialRequest is the body sent to PUT /v2/credentials/{reference}.
// The name field is immutable; only description and the typed sub-object can change.
type UpdateCredentialRequest struct {
	Name                           string                          `json:"name"`
	Description                    string                          `json:"description"`
	Basic                          *BasicAuth                      `json:"basic,omitempty"`
	CloudConnector                 *CloudConnector                 `json:"cloudConnector,omitempty"`
	WebhookToken                   *WebhookToken                   `json:"webhookToken,omitempty"`
	ContainerRegistryConfiguration *ContainerRegistryConfiguration `json:"containerRegistryConfiguration,omitempty"`
	KubernetesConfiguration        *KubernetesConfiguration        `json:"kubernetesConfiguration,omitempty"`
}

// PatchCredentialRequest is the body sent to PATCH /v2/credentials/{reference}.
// Only non-nil fields are sent — omitempty ensures zero values are omitted.
// Use pointer fields so callers can distinguish "set to empty" from "not provided".
type PatchCredentialRequest struct {
	Name                           *string                         `json:"name,omitempty"`
	Description                    *string                         `json:"description,omitempty"`
	Basic                          *BasicAuth                      `json:"basic,omitempty"`
	CloudConnector                 *CloudConnector                 `json:"cloudConnector,omitempty"`
	WebhookToken                   *WebhookToken                   `json:"webhookToken,omitempty"`
	ContainerRegistryConfiguration *ContainerRegistryConfiguration `json:"containerRegistryConfiguration,omitempty"`
	KubernetesConfiguration        *KubernetesConfiguration        `json:"kubernetesConfiguration,omitempty"`
}

// BasicAuth is the write payload for the basic-auth credential sub-type.
type BasicAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// CloudConnectorModel is the read-response sub-object for cloud-connector credentials.
type CloudConnectorModel struct {
	LocationID string `json:"locationId"`
}

// CloudConnector is the write payload for the cloud-connector credential sub-type.
type CloudConnector struct {
	LocationID string `json:"locationId"`
}

// WebhookToken is the write payload for the webhook-secret credential sub-type.
// The token is write-only — the API never returns it on read.
type WebhookToken struct {
	Token string `json:"token"`
}

// ContainerRegistryConfiguration is the write payload for the container-registry credential sub-type.
// The content is write-only — the API never returns it on read.
type ContainerRegistryConfiguration struct {
	Content string `json:"content"`
}

// KubernetesConfiguration is the write payload for the kubernetes-config credential sub-type.
// The content is write-only — the API never returns it on read.
type KubernetesConfiguration struct {
	Content string `json:"content"`
}

// CredentialListResponse is the envelope returned by GET /v2/credentials.
type CredentialListResponse struct {
	Embedded *CredentialListEmbedded `json:"_embedded,omitempty"`
}

type CredentialListEmbedded struct {
	Credentials []Credential `json:"credentials"`
}
