// btpservices/provider/cicd/jobs/list_resource_trigger.go

package cicdjobs

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	cicdclient "github.com/SAP/terraform-provider-btp-services/internal/cicd/client"
	"github.com/SAP/terraform-provider-btp-services/internal/shared"
)

var _ list.ListResourceWithConfigure = &triggerListResource{}

type triggerListResource struct {
	cli *cicdclient.CicdClientFacade
}

func NewTriggerListResource() list.ListResource {
	return &triggerListResource{}
}

func (r *triggerListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_cicd_trigger", req.ProviderTypeName)
}

func (r *triggerListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
			"A cicd{} block must be configured in the provider to use CI/CD resources.",
		)
		return
	}
	r.cli = clients.Cicd
}

func (r *triggerListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{
		MarkdownDescription: "Lists all triggers for a CI/CD job in the SAP BTP CI/CD service.",
		Attributes: map[string]listschema.Attribute{
			"job": listschema.StringAttribute{
				MarkdownDescription: "Name or ID of the CI/CD job whose triggers to list.",
				Required:            true,
			},
		},
	}
}

type triggerListConfigModel struct {
	Job types.String `tfsdk:"job"`
}

func (r *triggerListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {

	var config triggerListConfigModel
	diags := req.Config.Get(ctx, &config)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	triggers, err := r.cli.Jobs.ListTriggers(ctx, config.Job.ValueString())
	if err != nil {
		var d diag.Diagnostics
		d.AddError(
			"API Error Listing CI/CD Triggers",
			fmt.Sprintf("Failed to list triggers for job %q: %s", config.Job.ValueString(), err),
		)
		stream.Results = list.ListResultsStreamDiagnostics(d)
		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, trigger := range triggers {
			result := req.NewListResult(ctx)
			result.Identity.SetAttribute(ctx, path.Root("job"), config.Job)
			result.Identity.SetAttribute(ctx, path.Root("id"), trigger.ID)

			if req.IncludeResource {
				model := triggerResourceValueFrom(config.Job.ValueString(), trigger)
				result.Diagnostics.Append(result.Resource.Set(ctx, model)...)
			}

			if !push(result) {
				return
			}
		}
	}
}
