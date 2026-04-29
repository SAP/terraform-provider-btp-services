// internal/shared/provider_clients.go

package shared

import (
	cicdclient "github.com/SAP/terraform-provider-sap-btp-services/internal/cicd/client"
)

// ProviderClients is stored as ProviderData after provider Configure().
// Resources and data sources type-assert directly:
//
//	clients := req.ProviderData.(*shared.ProviderClients)
type ProviderClients struct {
	Cicd *cicdclient.CicdClientFacade
}
