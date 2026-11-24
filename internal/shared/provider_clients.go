// internal/shared/provider_clients.go

package shared

import (
	cicdclient "github.com/SAP/terraform-provider-btp-services/internal/cicd/client"
	ctmsclient "github.com/SAP/terraform-provider-btp-services/internal/ctms/client"
)

// ProviderClients is stored as ProviderData after provider Configure().
// Resources and data sources type-assert directly:
//
//	clients := req.ProviderData.(*shared.ProviderClients)
type ProviderClients struct {
	Cicd      *cicdclient.CicdClientFacade
	Ctms      *ctmsclient.CtmsClientFacade
	CtmsScope CtmsScope
}

// CtmsScope records the service endpoint used in resource identities.
// Credentials determine the tenant; the endpoint may be shared across tenants.
type CtmsScope struct {
	Endpoint string
}
