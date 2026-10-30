package ctmsclient

import "time"

// CtmsClientConfig is resolved by the provider before constructing its client.
type CtmsClientConfig struct {
	// Endpoint is the service root without /v2, for example
	// https://transport-service-app-backend.ts.cfapps.eu10.hana.ondemand.com.
	Endpoint string
	// TokenURL is the complete OAuth2 client-credentials token endpoint.
	TokenURL     string
	ClientID     string
	ClientSecret string
	// Timeout defaults to 60 seconds when zero.
	Timeout time.Duration
	// UserAgent identifies the provider and Terraform versions on API requests.
	UserAgent string
	// NamedUser is displayed in CTMS for node/route mutations. Empty omits the
	// header, using the authentication user. Maximum length is 60 characters.
	// It is provider-level request metadata, never resource state.
	NamedUser string
}

const DefaultTimeout = 60 * time.Second
