package ctmsclient

import "net/http"

// CtmsClientFacade is the concrete entry point for CTMS landscape operations.
// Each child facade shares transport and authentication, as in the CI/CD client.
type CtmsClientFacade struct {
	httpClient   *ctmsHTTPClient
	Nodes        nodesFacade
	Routes       routesFacade
	Destinations destinationsFacade
	UI           uiFacade
}

// NewCtmsClientFacade constructs an OAuth2 client using the default retry transport.
func NewCtmsClientFacade(cfg CtmsClientConfig) *CtmsClientFacade {
	return newFacade(newCtmsHTTPClient(cfg))
}

// NewCtmsClientFacadeWithHTTP injects an HTTP client, for example for VCR tests.
// The supplied client's timeout and transport are used unchanged.
func NewCtmsClientFacadeWithHTTP(cfg CtmsClientConfig, client *http.Client) *CtmsClientFacade {
	return newFacade(newCtmsHTTPClientWithHTTP(cfg, client))
}

func newFacade(hc *ctmsHTTPClient) *CtmsClientFacade {
	return &CtmsClientFacade{
		httpClient:   hc,
		Nodes:        newNodesFacade(hc),
		Routes:       newRoutesFacade(hc),
		Destinations: newDestinationsFacade(hc),
		UI:           newUIFacade(hc),
	}
}
