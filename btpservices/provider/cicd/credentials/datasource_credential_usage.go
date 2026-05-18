package cicdcredentials

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"

	cicdclient "github.com/SAP/terraform-provider-sap-btp-services/internal/cicd/client"
	cicdmodels "github.com/SAP/terraform-provider-sap-btp-services/internal/cicd/models"
	"github.com/SAP/terraform-provider-sap-btp-services/internal/shared"
)

var _ datasource.DataSource = &credentialUsageDataSource{}
var _ datasource.DataSourceWithConfigure = &credentialUsageDataSource{}

func NewCredentialUsageDataSource() datasource.DataSource {
	return &credentialUsageDataSource{}
}

type credentialUsageDataSource struct {
	cli *cicdclient.CicdClientFacade
}

func (d *credentialUsageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_cicd_credential_usage", req.ProviderTypeName)
}

func (d *credentialUsageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all jobs and repositories that reference a given credential in the SAP BTP CI/CD service.",
		Attributes: map[string]schema.Attribute{
			"credential": schema.StringAttribute{
				MarkdownDescription: "Name or ID of the credential whose usages to retrieve.",
				Required:            true,
			},
			"usertype": schema.StringAttribute{
				MarkdownDescription: "Filter results by user type. Allowed values: `job`, `repository`. Omit to return all usages.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("job", "repository"),
				},
			},
			"usages": schema.ListNestedAttribute{
				MarkdownDescription: "The list of usages for the credential.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Unique identifier of the job or repository.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the job or repository.",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Type of the usage entry: `job` or `repository`.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *credentialUsageDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *credentialUsageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config credentialUsageDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	usertype := ""
	if !config.UserType.IsNull() && !config.UserType.IsUnknown() {
		usertype = config.UserType.ValueString()
	}
	usages, err := d.cli.Credentials.GetUsages(ctx, config.Credential.ValueString(), usertype)
	if err != nil {
		if cicdmodels.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Credential Not Found",
				fmt.Sprintf("No credential found with reference %q.", config.Credential.ValueString()),
			)
			return
		}
		resp.Diagnostics.AddError("Error Reading Credential Usages", err.Error())
		return
	}

	usagesList, diags := credentialUsageDSItemsFrom(usages)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := credentialUsageDSModel{
		Credential: config.Credential,
		UserType:   config.UserType,
		Usages:     usagesList,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
