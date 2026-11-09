// btpservices/provider/provider.go

package provider

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/cicd"
	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms"
	"github.com/SAP/terraform-provider-btp-services/internal/shared"
	"github.com/SAP/terraform-provider-btp-services/internal/version"
)

var _ provider.Provider = &btpServicesProvider{}
var _ provider.ProviderWithListResources = &btpServicesProvider{}
var _ provider.ProviderWithActions = &btpServicesProvider{}

func New() func() provider.Provider {
	return func() provider.Provider {
		return &btpServicesProvider{}
	}
}

func NewWithClients(clients *shared.ProviderClients) provider.Provider {
	return &btpServicesProvider{prebuiltClients: clients}
}

type btpServicesProvider struct {
	prebuiltClients *shared.ProviderClients
}

type providerModel struct {
	Cicd *cicdProviderModel `tfsdk:"cicd"`
	Ctms *ctmsProviderModel `tfsdk:"ctms"`
}

func (p *btpServicesProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "btpservice"
}

func (p *btpServicesProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages resources across SAP Business Technology Platform services.",
		Blocks: map[string]schema.Block{
			"ctms": ctmsProviderBlock(),
			"cicd": cicdProviderBlock(),
		},
	}
}

func (p *btpServicesProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	if p.prebuiltClients != nil {
		resp.ResourceData = p.prebuiltClients
		resp.DataSourceData = p.prebuiltClients
		resp.ListResourceData = p.prebuiltClients
		resp.ActionData = p.prebuiltClients
		return
	}

	var data providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clients := &shared.ProviderClients{}

	if data.Cicd != nil {
		configureCICD(ctx, data.Cicd, req.TerraformVersion, clients, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if data.Ctms != nil {
		configureCTMS(ctx, data.Ctms, req.TerraformVersion, clients, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.ResourceData = clients
	resp.DataSourceData = clients
	resp.ListResourceData = clients
	resp.ActionData = clients
}

// servicePackages is the single service registry.
// To add a service: implement the pattern and append here.
func servicePackages() []interface {
	Resources(context.Context) []func() resource.Resource
	DataSources(context.Context) []func() datasource.DataSource
	ListResources(context.Context) []func() list.ListResource
	Actions(context.Context) []func() action.Action
} {
	return []interface {
		Resources(context.Context) []func() resource.Resource
		DataSources(context.Context) []func() datasource.DataSource
		ListResources(context.Context) []func() list.ListResource
		Actions(context.Context) []func() action.Action
	}{
		cicd.ServicePackage{},
		ctms.ServicePackage{},
	}
}

func (p *btpServicesProvider) Resources(ctx context.Context) []func() resource.Resource {
	var all []func() resource.Resource
	for _, pkg := range servicePackages() {
		all = append(all, pkg.Resources(ctx)...)
	}
	return all
}

func (p *btpServicesProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	var all []func() datasource.DataSource
	for _, pkg := range servicePackages() {
		all = append(all, pkg.DataSources(ctx)...)
	}
	return all
}

func (p *btpServicesProvider) ListResources(ctx context.Context) []func() list.ListResource {
	var all []func() list.ListResource
	for _, pkg := range servicePackages() {
		all = append(all, pkg.ListResources(ctx)...)
	}
	return all
}

func (p *btpServicesProvider) Actions(ctx context.Context) []func() action.Action {
	var all []func() action.Action
	for _, pkg := range servicePackages() {
		all = append(all, pkg.Actions(ctx)...)
	}
	return all
}

// buildUserAgent constructs the User-Agent header value:
//
//	Terraform/<tfVersion> terraform-provider-btp-services/<providerVersion> [<custom>]
//
// The optional custom suffix is read from BTP_SERVICE_USER_AGENT_SUFFIX.
func buildUserAgent(tfVersion string) string {
	base := fmt.Sprintf("Terraform/%s terraform-provider-btp-services/%s", tfVersion, version.ProviderVersion)
	if custom := strings.TrimSpace(os.Getenv("BTP_SERVICE_USER_AGENT_SUFFIX")); custom != "" {
		return base + " " + custom
	}
	return base
}
