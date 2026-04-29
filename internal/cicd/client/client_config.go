// internal/services/cicd/client/client_config.go

package cicdclient

import "time"

// CicdClientConfig holds the configuration required to build a CicdClientFacade.
// Values are resolved from HCL provider block fields or environment variables
// by the provider's Configure() function before this struct is constructed.
type CicdClientConfig struct {
	// Endpoint is the CI/CD service base URL, e.g.
	// https://cicd-service.cfapps.eu12.hana.ondemand.com
	Endpoint string

	// TokenURL is the OAuth2 client-credentials token endpoint, e.g.
	// https://<identityzone>.authentication.eu12.hana.ondemand.com/oauth/token
	TokenURL string

	// ClientID and ClientSecret are the OAuth2 client credentials used to
	// obtain a bearer token via the client_credentials grant.
	ClientID     string
	ClientSecret string

	// Timeout is the HTTP client timeout per request. Defaults to 60s when zero.
	Timeout time.Duration
}

// DefaultTimeout is used when CicdClientConfig.Timeout is not set.
const DefaultTimeout = 60 * time.Second
