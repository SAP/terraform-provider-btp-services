package ctms

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	ctmsdestinations "github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/destinations"
	ctmsnodes "github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/nodes"
	ctmsroutes "github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/routes"
	ctmsui "github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/ui"
)

// ServicePackage registers the landscape operations exposed by the CTMS client.
type ServicePackage struct{}

func (ServicePackage) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		ctmsnodes.NewNodeResource,
		ctmsroutes.NewRouteResource,
	}
}

func (ServicePackage) DataSources(context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		ctmsnodes.NewNodeDataSource,
		ctmsnodes.NewNodesDataSource,
		ctmsroutes.NewRouteDataSource,
		ctmsroutes.NewRoutesDataSource,
		ctmsdestinations.NewDestinationDataSource,
		ctmsdestinations.NewDestinationsDataSource,
		ctmsui.NewUIURLDataSource,
		ctmsui.NewTransportRequestLogsURLDataSource,
	}
}

func (ServicePackage) ListResources(context.Context) []func() list.ListResource {
	return []func() list.ListResource{
		ctmsnodes.NewNodeListResource,
		ctmsroutes.NewRouteListResource,
	}
}

func (ServicePackage) Actions(context.Context) []func() action.Action {
	return nil
}
