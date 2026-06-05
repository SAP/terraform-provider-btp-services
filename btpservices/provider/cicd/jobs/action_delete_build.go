// btpservices/provider/cicd/jobs/action_delete_build.go

package cicdjobs

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	cicdclient "github.com/SAP/terraform-provider-btp-services/internal/cicd/client"
	cicdmodels "github.com/SAP/terraform-provider-btp-services/internal/cicd/models"
	"github.com/SAP/terraform-provider-btp-services/internal/shared"
)

var _ action.Action = &deleteBuildAction{}
var _ action.ActionWithConfigure = &deleteBuildAction{}

// NewDeleteBuildAction is the constructor for btpservice_cicd_delete_build.
func NewDeleteBuildAction() action.Action {
	return &deleteBuildAction{}
}

type deleteBuildAction struct {
	cli *cicdclient.CicdClientFacade
}

// deleteBuildModel is the configuration model for the btpservice_cicd_delete_build action.
type deleteBuildModel struct {
	Job   types.String `tfsdk:"job"`
	Build types.String `tfsdk:"build"`
}

func (a *deleteBuildAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_cicd_delete_build", req.ProviderTypeName)
}

func (a *deleteBuildAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Deletes a specific CI/CD build.",
		Attributes: map[string]schema.Attribute{
			"job": schema.StringAttribute{
				MarkdownDescription: "Name or ID of the CI/CD job that owns the build.",
				Required:            true,
			},
			"build": schema.StringAttribute{
				MarkdownDescription: "Build sequence number or ID to delete.",
				Required:            true,
			},
		},
	}
}

func (a *deleteBuildAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
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
			"A cicd{} block must be configured in the provider to use CI/CD actions.",
		)
		return
	}
	a.cli = clients.Cicd
}

func (a *deleteBuildAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config deleteBuildModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := a.cli.Builds.Delete(ctx, config.Job.ValueString(), config.Build.ValueString()); err != nil {
		if cicdmodels.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Build Not Found",
				fmt.Sprintf("No build %q found for job %q.", config.Build.ValueString(), config.Job.ValueString()),
			)
			return
		}
		resp.Diagnostics.AddError("Error Deleting Build", err.Error())
		return
	}
}
