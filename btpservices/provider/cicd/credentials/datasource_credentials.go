// btpservices/provider/cicd/credentials/datasource_credentials.go

package cicdcredentials

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	cicdclient "github.com/SAP/terraform-provider-btp-services/internal/cicd/client"
	"github.com/SAP/terraform-provider-btp-services/internal/shared"
)

var _ datasource.DataSource = &credentialsDataSource{}
var _ datasource.DataSourceWithConfigure = &credentialsDataSource{}

func NewCredentialsDataSource() datasource.DataSource {
	return &credentialsDataSource{}
}

type credentialsDataSource struct {
	cli *cicdclient.CicdClientFacade
}

func (d *credentialsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_cicd_credentials", req.ProviderTypeName)
}

func (d *credentialsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all credentials in the SAP BTP CI/CD service. Passwords are never returned by the API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the resource (assigned by the API).",
				Computed:            true,
			},
			"values": schema.ListNestedAttribute{
				MarkdownDescription: "The list of credentials.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The unique identifier of the credential.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the credential.",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Human-readable description of the credential.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *credentialsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	clients, ok := req.ProviderData.(*shared.ProviderClients)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			fmt.Sprintf("Expected *shared.ProviderClients, got: %T", req.ProviderData),
		)
		return
	}

	if clients.Cicd == nil {
		resp.Diagnostics.AddError(
			"Missing CI/CD Configuration",
			"A cicd{} block must be configured in the provider to use CI/CD data sources.",
		)
		return
	}
	d.cli = clients.Cicd
}

func (d *credentialsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	list, err := d.cli.Credentials.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Listing Credentials", err.Error())
		return
	}

	state := credentialsDSModel{
		ID:     types.StringValue("credentials"),
		Values: credentialsDSItemsFrom(list),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
