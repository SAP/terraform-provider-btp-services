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

// CtmsScope identifies the landscape associated with the configured credentials.
// SubaccountID is caller supplied; the landscape API does not return this scope.
type CtmsScope struct {
	Endpoint     string
	SubaccountID string
}
