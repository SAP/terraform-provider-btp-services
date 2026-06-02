// btpservices/provider/cicd/credentials/datasource_basic_auth.go

package cicdcredentials

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"

	cicdclient "github.com/SAP/terraform-provider-btp-services/internal/cicd/client"
	cicdmodels "github.com/SAP/terraform-provider-btp-services/internal/cicd/models"
	"github.com/SAP/terraform-provider-btp-services/internal/shared"
)

var _ datasource.DataSource = &basicAuthDataSource{}
var _ datasource.DataSourceWithConfigure = &basicAuthDataSource{}

// NewBasicAuthDataSource is the constructor exported to service_package.go.
func NewCredentialDataSource() datasource.DataSource {
	return &basicAuthDataSource{}
}

type basicAuthDataSource struct {
	cli *cicdclient.CicdClientFacade
}

func (d *basicAuthDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_cicd_credential", req.ProviderTypeName)
}

func (d *basicAuthDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads credential from the SAP BTP CI/CD service. The password is never returned by the API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the credential. Exactly one of `id` or `name` must be set.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRoot("name"),
					}...),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the credential. Exactly one of `id` or `name` must be set.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRoot("id"),
					}...),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Human-readable description of the credential.",
				Computed:            true,
			},
		},
	}
}

func (d *basicAuthDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *basicAuthDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config basicAuthDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var reference string
	if !config.ID.IsNull() && !config.ID.IsUnknown() {
		reference = config.ID.ValueString()
	} else {
		reference = config.Name.ValueString()
	}

	result, err := d.cli.Credentials.Get(ctx, reference)
	if err != nil {
		if cicdmodels.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Credential Not Found",
				fmt.Sprintf("No credential found with reference %q.", reference),
			)
			return
		}

		resp.Diagnostics.AddError("Error Reading Credential", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, basicCredsDSValueFrom(*result))...)
}
