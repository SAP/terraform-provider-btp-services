// internal/cicd/models/credential.go

package cicdmodels

// Credential is the API response model for a single credential.
// Sensitive fields (password, text, key, etc.) are never returned by the API on read.
// The API always returns the type key (e.g. "webhookToken":{}) even when the
// content is write-only, so the presence of a non-nil pointer identifies the type.
type Credential struct {
	ID          string `json:"id"`
	APIVersion  string `json:"apiVersion"`
	Name        string `json:"name"`
	Description string `json:"description"`

	// Only the sub-object matching the credential type is populated on read.
	Basic                            *BasicAuthModel                        `json:"basic,omitempty"`
	BasicForCustomIdP                *BasicForCustomIdPModel                `json:"basicForCustomIdP,omitempty"`
	CertificateBasedAuthForCustomIdP *CertificateBasedAuthForCustomIdPModel `json:"certificateBasedAuthForCustomIdP,omitempty"`

	// CloudConnector is populated only when the credential is of Cloud Connector type.
	// locationId is readable on GET.
	CloudConnector *CloudConnectorModel `json:"cloudConnector,omitempty"`

	// WebhookToken is present (as an empty object) when the credential is of Webhook Secret type.
	// The token itself is write-only and never returned.
	WebhookToken *WebhookTokenModel `json:"webhookToken,omitempty"`

	// ContainerRegistryConfiguration is present (as an empty object) when the credential is of
	// Container Registry type. The content is write-only and never returned.
	ContainerRegistryConfiguration *ContainerRegistryConfigurationModel `json:"containerRegistryConfiguration,omitempty"`

	// KubernetesConfiguration is present (as an empty object) when the credential is of
	// Kubernetes Config type. The content is write-only and never returned.
	KubernetesConfiguration *KubernetesConfigurationModel `json:"kubernetesConfiguration,omitempty"`

	// SecretText is present (as an empty object) when the credential is of Secret Text type.
	// The text is write-only and never returned.
	SecretText *SecretTextModel `json:"secretText,omitempty"`

	// ServiceKey is present (as an empty object) when the credential is of Service Key type.
	// The key is write-only and never returned.
	ServiceKey *ServiceKeyModel `json:"serviceKey,omitempty"`
}

// BasicAuthModel is the read-response sub-object for basic-auth credentials.
// The password field is intentionally absent — the API never returns it.
type BasicAuthModel struct {
	Username string `json:"username"`
}

// BasicForCustomIdPModel is the read-response sub-object for basic-auth custom-IdP credentials.
// The password field is never returned by the API.
type BasicForCustomIdPModel struct {
	Username string `json:"username"`
	Origin   string `json:"origin,omitempty"`
}

// CertificateBasedAuthForCustomIdPModel is the read-response sub-object for
// certificate-based custom-IdP credentials. All fields are readable.
type CertificateBasedAuthForCustomIdPModel struct {
	EmailAddress string `json:"emailAddress,omitempty"`
	Hostname     string `json:"hostname,omitempty"`
	Origin       string `json:"origin,omitempty"`
}

// SecretTextModel is the read-response sub-object for secret-text credentials.
// The text field is write-only — the API never returns it on read.
type SecretTextModel struct{}

// ServiceKeyModel is the read-response sub-object for service-key credentials.
// The key field is write-only — the API never returns it on read.
type ServiceKeyModel struct{}

// WebhookTokenModel is the read-response sub-object for webhook-token credentials.
// The token field is write-only — the API never returns it on read.
type WebhookTokenModel struct{}

// ContainerRegistryConfigurationModel is the read-response sub-object for container-registry credentials.
// The content field is write-only — the API never returns it on read.
type ContainerRegistryConfigurationModel struct{}

// KubernetesConfigurationModel is the read-response sub-object for kubernetes-config credentials.
// The content field is write-only — the API never returns it on read.
type KubernetesConfigurationModel struct{}

// CreateCredentialRequest is the body sent to POST /v2/credentials.
// Exactly one of the typed sub-objects must be set.
type CreateCredentialRequest struct {
	Name                             string                            `json:"name"`
	Description                      string                            `json:"description"`
	Basic                            *BasicAuth                        `json:"basic,omitempty"`
	CloudConnector                   *CloudConnector                   `json:"cloudConnector,omitempty"`
	WebhookToken                     *WebhookToken                     `json:"webhookToken,omitempty"`
	ContainerRegistryConfiguration   *ContainerRegistryConfiguration   `json:"containerRegistryConfiguration,omitempty"`
	KubernetesConfiguration          *KubernetesConfiguration          `json:"kubernetesConfiguration,omitempty"`
	BasicForCustomIdP                *BasicForCustomIdP                `json:"basicForCustomIdP,omitempty"`
	CertificateBasedAuthForCustomIdP *CertificateBasedAuthForCustomIdP `json:"certificateBasedAuthForCustomIdP,omitempty"`
	ServiceKey                       *ServiceKey                       `json:"serviceKey,omitempty"`
	SecretText                       *SecretText                       `json:"secretText,omitempty"`
}

// UpdateCredentialRequest is the body sent to PUT /v2/credentials/{reference}.
type UpdateCredentialRequest struct {
	Name                             string                            `json:"name"`
	Description                      string                            `json:"description"`
	Basic                            *BasicAuth                        `json:"basic,omitempty"`
	CloudConnector                   *CloudConnector                   `json:"cloudConnector,omitempty"`
	WebhookToken                     *WebhookToken                     `json:"webhookToken,omitempty"`
	ContainerRegistryConfiguration   *ContainerRegistryConfiguration   `json:"containerRegistryConfiguration,omitempty"`
	KubernetesConfiguration          *KubernetesConfiguration          `json:"kubernetesConfiguration,omitempty"`
	BasicForCustomIdP                *BasicForCustomIdP                `json:"basicForCustomIdP,omitempty"`
	CertificateBasedAuthForCustomIdP *CertificateBasedAuthForCustomIdP `json:"certificateBasedAuthForCustomIdP,omitempty"`
	ServiceKey                       *ServiceKey                       `json:"serviceKey,omitempty"`
	SecretText                       *SecretText                       `json:"secretText,omitempty"`
}

// PatchCredentialRequest is the body sent to PATCH /v2/credentials/{reference}.
// Only non-nil fields are sent — omitempty ensures zero values are omitted.
type PatchCredentialRequest struct {
	Name                             *string                           `json:"name,omitempty"`
	Description                      *string                           `json:"description,omitempty"`
	Basic                            *BasicAuth                        `json:"basic,omitempty"`
	CloudConnector                   *CloudConnector                   `json:"cloudConnector,omitempty"`
	WebhookToken                     *WebhookToken                     `json:"webhookToken,omitempty"`
	ContainerRegistryConfiguration   *ContainerRegistryConfiguration   `json:"containerRegistryConfiguration,omitempty"`
	KubernetesConfiguration          *KubernetesConfiguration          `json:"kubernetesConfiguration,omitempty"`
	BasicForCustomIdP                *BasicForCustomIdP                `json:"basicForCustomIdP,omitempty"`
	CertificateBasedAuthForCustomIdP *CertificateBasedAuthForCustomIdP `json:"certificateBasedAuthForCustomIdP,omitempty"`
	ServiceKey                       *ServiceKey                       `json:"serviceKey,omitempty"`
	SecretText                       *SecretText                       `json:"secretText,omitempty"`
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

// BasicForCustomIdP is the write payload for the basic-auth custom-IdP credential sub-type.
// Origin is the custom identity provider's origin key (e.g. "custom-platform").
type BasicForCustomIdP struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Origin   string `json:"origin"`
}

// CertificateBasedAuthForCustomIdP is the write payload for the certificate-based
// custom-IdP credential sub-type. Authenticates using IAS tenant hostname + origin.
type CertificateBasedAuthForCustomIdP struct {
	EmailAddress string `json:"emailAddress"`
	Hostname     string `json:"hostname"`
	Origin       string `json:"origin"`
}

// ServiceKey is the write payload for the service-key credential sub-type.
// Key must be valid JSON (SAP BTP service binding key).
type ServiceKey struct {
	Key string `json:"key"`
}

// SecretText is the write payload for the secret-text credential sub-type.
type SecretText struct {
	Text string `json:"text"`
}

// CredentialListResponse is the envelope returned by GET /v2/credentials.
type CredentialListResponse struct {
	Embedded *CredentialListEmbedded `json:"_embedded,omitempty"`
}

type CredentialListEmbedded struct {
	Credentials []Credential `json:"credentials"`
}
