// internal/cicd/client/facade.go

package cicdclient

import (
	"fmt"
	"net/http"
)

// CicdClientFacade is the entry point for all CI/CD API operations.
// Resources store a *CicdClientFacade directly — no interface, mirrors BTP's *btpcli.ClientFacade.
type CicdClientFacade struct {
	httpClient    *cicdHTTPClient
	Credentials   credentialsFacade
	Repositories  repositoriesFacade
	Jobs          jobsFacade
	AllowedSpaces allowedSpacesFacade
	Builds        buildsFacade
}

// NewCicdClientFacade constructs a facade using the OAuth2 client-credentials flow
// to obtain bearer tokens automatically. Returns an error if the configuration
// contains non-HTTPS URLs.
func NewCicdClientFacade(cfg CicdClientConfig) (*CicdClientFacade, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid CI/CD client configuration: %w", err)
	}
	hc := newCicdHTTPClient(cfg)
	return &CicdClientFacade{
		httpClient:    hc,
		Credentials:   newCredentialsFacade(hc),
		Repositories:  newRepositoriesFacade(hc),
		Jobs:          newJobsFacade(hc),
		AllowedSpaces: newAllowedSpacesFacade(hc),
		Builds:        newBuildsFacade(hc),
	}, nil
}

// NewCicdClientFacadeWithHTTP injects a custom *http.Client — used by VCR acceptance tests.
func NewCicdClientFacadeWithHTTP(cfg CicdClientConfig, httpClient *http.Client) *CicdClientFacade {
	hc := newCicdHTTPClientWithHTTP(cfg, httpClient)
	return &CicdClientFacade{
		httpClient:    hc,
		Credentials:   newCredentialsFacade(hc),
		Repositories:  newRepositoriesFacade(hc),
		Jobs:          newJobsFacade(hc),
		AllowedSpaces: newAllowedSpacesFacade(hc),
		Builds:        newBuildsFacade(hc),
	}
}
