// internal/cicd/client/facade.go

package cicdclient

import "net/http"

// CicdClientFacade is the entry point for all CI/CD API operations.
// Resources store a *CicdClientFacade directly — no interface, mirrors BTP's *btpcli.ClientFacade.
type CicdClientFacade struct {
	httpClient  *cicdHTTPClient
	Credentials credentialsFacade
}

// NewCicdClientFacade constructs a facade using the OAuth2 client-credentials flow
// to obtain bearer tokens automatically.
func NewCicdClientFacade(cfg CicdClientConfig) *CicdClientFacade {
	hc := newCicdHTTPClient(cfg)
	return &CicdClientFacade{
		httpClient:  hc,
		Credentials: newCredentialsFacade(hc),
	}
}

// NewCicdClientFacadeWithHTTP injects a custom *http.Client — used by VCR acceptance tests.
func NewCicdClientFacadeWithHTTP(cfg CicdClientConfig, httpClient *http.Client) *CicdClientFacade {
	hc := newCicdHTTPClientWithHTTP(cfg, httpClient)
	return &CicdClientFacade{
		httpClient:  hc,
		Credentials: newCredentialsFacade(hc),
	}
}
