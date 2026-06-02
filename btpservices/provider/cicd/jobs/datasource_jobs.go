// btpservices/provider/cicd/jobs/datasource_jobs.go

package cicdjobs

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	cicdclient "github.com/SAP/terraform-provider-btp-services/internal/cicd/client"
	"github.com/SAP/terraform-provider-btp-services/internal/shared"
)

var _ datasource.DataSource = &jobsDataSource{}
var _ datasource.DataSourceWithConfigure = &jobsDataSource{}

func NewJobsDataSource() datasource.DataSource {
	return &jobsDataSource{}
}

type jobsDataSource struct {
	cli *cicdclient.CicdClientFacade
}

func (d *jobsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_cicd_jobs", req.ProviderTypeName)
}

func (d *jobsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all jobs in the SAP BTP CI/CD service.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the data source.",
				Computed:            true,
			},
			"values": schema.ListNestedAttribute{
				MarkdownDescription: "The list of jobs.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: jobsDSItemAttributes(),
				},
			},
		},
	}
}

// jobsDSItemAttributes returns the attribute map for each job item in the values list.
func jobsDSItemAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "Unique identifier of the job.",
			Computed:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the job.",
			Computed:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "Human-readable description of the job.",
			Computed:            true,
		},
		"active": schema.BoolAttribute{
			MarkdownDescription: "Whether the job is active.",
			Computed:            true,
		},
		"pipeline": schema.StringAttribute{
			MarkdownDescription: "Pipeline type of the job (e.g. `cf-env`, `cpi`).",
			Computed:            true,
		},
		"pipeline_version": schema.StringAttribute{
			MarkdownDescription: "Version of the pipeline type.",
			Computed:            true,
		},
		"pipeline_parameters": schema.StringAttribute{
			MarkdownDescription: "Pipeline parameters serialised as a canonical YAML string.",
			Computed:            true,
		},
		"build_retention_days": schema.Int64Attribute{
			MarkdownDescription: "Number of days build artifacts are retained.",
			Computed:            true,
		},
		"max_builds_to_keep": schema.Int64Attribute{
			MarkdownDescription: "Maximum number of builds retained for this job.",
			Computed:            true,
		},
		"branch": schema.StringAttribute{
			MarkdownDescription: "Branch pattern the job is executed for.",
			Computed:            true,
		},
		"repository_id": schema.StringAttribute{
			MarkdownDescription: "ID of the source repository used by the job.",
			Computed:            true,
		},
		"notification_configuration": schema.SingleNestedAttribute{
			MarkdownDescription: "Notification settings for the job.",
			Computed:            true,
			Attributes: map[string]schema.Attribute{
				"ans": schema.SingleNestedAttribute{
					MarkdownDescription: "SAP Alert Notification Service (ANS) settings.",
					Computed:            true,
					Attributes: map[string]schema.Attribute{
						"active": schema.BoolAttribute{
							MarkdownDescription: "Whether ANS notifications are active.",
							Computed:            true,
						},
						"credential_id": schema.StringAttribute{
							MarkdownDescription: "ID of the ANS credential to use.",
							Computed:            true,
						},
						"custom_tag": schema.StringAttribute{
							MarkdownDescription: "Optional custom tag added to ANS notifications.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *jobsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *jobsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	list, err := d.cli.Jobs.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Listing Jobs", err.Error())
		return
	}

	state := jobsDSModel{
		ID: types.StringValue("jobs"),
	}
	values, diags := jobsDSItemsFrom(ctx, list)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Values = values

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
