// btpservices/provider/cicd/jobs/datasource_triggers.go

package cicdjobs

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	cicdclient "github.com/SAP/terraform-provider-sap-btp-services/internal/cicd/client"
	"github.com/SAP/terraform-provider-sap-btp-services/internal/shared"
)

var _ datasource.DataSource = &triggersDataSource{}
var _ datasource.DataSourceWithConfigure = &triggersDataSource{}

func NewTriggersDataSource() datasource.DataSource {
	return &triggersDataSource{}
}

type triggersDataSource struct {
	cli *cicdclient.CicdClientFacade
}

func (d *triggersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_cicd_triggers", req.ProviderTypeName)
}

func (d *triggersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all triggers for a CI/CD job in the SAP BTP CI/CD service.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier for this data source result.",
				Computed:            true,
			},
			"job": schema.StringAttribute{
				MarkdownDescription: "Name or ID of the CI/CD job whose triggers to list.",
				Required:            true,
			},
			"values": schema.ListNestedAttribute{
				MarkdownDescription: "The list of triggers for the specified job.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Unique identifier of the trigger.",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Trigger type (e.g. `timer`).",
							Computed:            true,
						},
						"timer": schema.SingleNestedAttribute{
							MarkdownDescription: "Timer schedule configuration. Present when `type` is `timer`.",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"branch": schema.StringAttribute{
									MarkdownDescription: "Branch to build on the timer schedule.",
									Computed:            true,
								},
								"cron": schema.StringAttribute{
									MarkdownDescription: "Cron expression defining the timer schedule.",
									Computed:            true,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *triggersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *triggersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config triggersDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	triggers, err := d.cli.Jobs.ListTriggers(ctx, config.Job.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Triggers",
			fmt.Sprintf("Failed to list triggers for job %q: %s", config.Job.ValueString(), err),
		)
		return
	}

	values := make([]triggerDSItem, len(triggers))
	for i, t := range triggers {
		values[i] = triggerDSItemFrom(t)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, triggersDSModel{
		ID:     types.StringValue("triggers/" + config.Job.ValueString()),
		Job:    config.Job,
		Values: values,
	})...)
}
