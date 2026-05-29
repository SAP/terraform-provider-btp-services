// btpservices/provider/cicd/jobs/datasource_trigger.go

package cicdjobs

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	cicdclient "github.com/SAP/terraform-provider-sap-btp-services/internal/cicd/client"
	cicdmodels "github.com/SAP/terraform-provider-sap-btp-services/internal/cicd/models"
	"github.com/SAP/terraform-provider-sap-btp-services/internal/shared"
)

var _ datasource.DataSource = &triggerDataSource{}
var _ datasource.DataSourceWithConfigure = &triggerDataSource{}

func NewTriggerDataSource() datasource.DataSource {
	return &triggerDataSource{}
}

type triggerDataSource struct {
	cli *cicdclient.CicdClientFacade
}

func (d *triggerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_cicd_trigger", req.ProviderTypeName)
}

func (d *triggerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a single trigger for a CI/CD job in the SAP BTP CI/CD service.",
		Attributes: map[string]schema.Attribute{
			"job": schema.StringAttribute{
				MarkdownDescription: "Name or ID of the CI/CD job that owns this trigger.",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the trigger.",
				Required:            true,
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
	}
}

func (d *triggerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *triggerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config triggerResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.cli.Jobs.GetTrigger(ctx, config.Job.ValueString(), config.ID.ValueString())
	if err != nil {
		if cicdmodels.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Trigger Not Found",
				fmt.Sprintf("No trigger found with ID %q in job %q.", config.ID.ValueString(), config.Job.ValueString()),
			)
			return
		}
		resp.Diagnostics.AddError("Error Reading Trigger", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, triggerResourceValueFrom(config.Job.ValueString(), *result))...)
}
