// internal/cicd/client/client_config.go

package cicdclient

import (
	"fmt"
	"net/url"
	"time"
)

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

	// UserAgent is sent as the User-Agent header on every request.
	// Populated by the provider's Configure() using the Terraform version and
	// provider version so the CI/CD service can identify the caller.
	UserAgent string
}

// DefaultTimeout is used when CicdClientConfig.Timeout is not set.
const DefaultTimeout = 60 * time.Second

// Validate returns an error if Endpoint or TokenURL are missing or do not use https.
func (c CicdClientConfig) Validate() error {
	if err := requireHTTPS("endpoint", c.Endpoint); err != nil {
		return err
	}
	return requireHTTPS("token_url", c.TokenURL)
}

func requireHTTPS(field, rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("%s must not be empty", field)
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("%s is not a valid URL: %w", field, err)
	}
	if u.Scheme != "https" {
		return fmt.Errorf("%s must use https (got %q)", field, u.Scheme)
	}
	return nil
}
