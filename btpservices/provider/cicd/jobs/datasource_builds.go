// btpservices/provider/cicd/jobs/datasource_builds.go

package cicdjobs

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	cicdclient "github.com/SAP/terraform-provider-btp-services/internal/cicd/client"
	cicdmodels "github.com/SAP/terraform-provider-btp-services/internal/cicd/models"
	"github.com/SAP/terraform-provider-btp-services/internal/shared"
)

var _ datasource.DataSource = &buildsDataSource{}
var _ datasource.DataSourceWithConfigure = &buildsDataSource{}

func NewBuildsDataSource() datasource.DataSource {
	return &buildsDataSource{}
}

type buildsDataSource struct {
	cli *cicdclient.CicdClientFacade
}

func (d *buildsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_cicd_builds", req.ProviderTypeName)
}

func (d *buildsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists builds for a CI/CD job filtered by `latest` (currently running or most recently triggered) or `latestFinished` (most recently completed).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the data source.",
				Computed:            true,
			},
			"job": schema.StringAttribute{
				MarkdownDescription: "Name or ID of the CI/CD job.",
				Required:            true,
			},
			"filter": schema.StringAttribute{
				MarkdownDescription: "Filter to apply. Must be `latest` (running or last triggered build) or `latestFinished` (last completed build).",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("latest", "latestFinished"),
				},
			},
			"builds": schema.ListNestedAttribute{
				MarkdownDescription: "The list of matching builds.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Technical immutable ID of the build.",
							Computed:            true,
						},
						"number": schema.Int64Attribute{
							MarkdownDescription: "Sequential build number.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *buildsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *buildsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config buildsDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	builds, err := d.cli.Builds.ListBuilds(ctx, config.Job.ValueString(), config.Filter.ValueString())
	if err != nil {
		if cicdmodels.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Job Not Found",
				fmt.Sprintf("No job found with reference %q.", config.Job.ValueString()),
			)
			return
		}
		resp.Diagnostics.AddError("Error Listing Builds", err.Error())
		return
	}

	state := buildsDSModel{
		ID:     types.StringValue(fmt.Sprintf("%s/%s", config.Job.ValueString(), config.Filter.ValueString())),
		Job:    config.Job,
		Filter: config.Filter,
		Builds: buildsDSItemsFrom(builds),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
