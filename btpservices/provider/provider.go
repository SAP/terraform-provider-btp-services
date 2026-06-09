// btpservices/provider/provider.go

package provider

import (
	"context"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/cicd"
	cicdclient "github.com/SAP/terraform-provider-btp-services/internal/cicd/client"
	"github.com/SAP/terraform-provider-btp-services/internal/shared"
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
}

type cicdProviderModel struct {
	Endpoint     types.String `tfsdk:"endpoint"`
	TokenURL     types.String `tfsdk:"token_url"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	Timeout      types.Int64  `tfsdk:"timeout"`
}

func (p *btpServicesProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "btpservice"
}

func (p *btpServicesProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages resources across SAP Business Technology Platform services.",
		Blocks: map[string]schema.Block{
			"cicd": schema.SingleNestedBlock{
				MarkdownDescription: "Configuration for the SAP BTP CI/CD service.",
				Attributes: map[string]schema.Attribute{
					"endpoint": schema.StringAttribute{
						MarkdownDescription: "CI/CD service base URL. Env: `BTP_CICD_ENDPOINT`.",
						Optional:            true,
					},
					"token_url": schema.StringAttribute{
						MarkdownDescription: "OAuth2 token endpoint. Env: `BTP_CICD_TOKEN_URL`.",
						Optional:            true,
					},
					"client_id": schema.StringAttribute{
						MarkdownDescription: "OAuth2 client ID. Env: `BTP_CICD_CLIENT_ID`.",
						Optional:            true,
					},
					"client_secret": schema.StringAttribute{
						MarkdownDescription: "OAuth2 client secret. Env: `BTP_CICD_CLIENT_SECRET`.",
						Optional:            true,
						Sensitive:           true,
					},
					"timeout": schema.Int64Attribute{
						MarkdownDescription: "HTTP request timeout in seconds. Defaults to 60.",
						Optional:            true,
					},
				},
			},
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
		cfg := cicdclient.CicdClientConfig{
			Endpoint:     resolveString(data.Cicd.Endpoint, "BTP_CICD_ENDPOINT"),
			TokenURL:     resolveString(data.Cicd.TokenURL, "BTP_CICD_TOKEN_URL"),
			ClientID:     resolveString(data.Cicd.ClientID, "BTP_CICD_CLIENT_ID"),
			ClientSecret: resolveString(data.Cicd.ClientSecret, "BTP_CICD_CLIENT_SECRET"),
		}
		if !data.Cicd.Timeout.IsNull() && !data.Cicd.Timeout.IsUnknown() {
			cfg.Timeout = time.Duration(data.Cicd.Timeout.ValueInt64()) * time.Second
		}
		clients.Cicd = cicdclient.NewCicdClientFacade(cfg)
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

func resolveString(v types.String, envKey string) string {
	if !v.IsNull() && !v.IsUnknown() {
		return v.ValueString()
	}
	return os.Getenv(envKey)
}
