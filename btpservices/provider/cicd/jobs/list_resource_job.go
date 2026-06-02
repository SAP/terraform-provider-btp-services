// btpservices/provider/cicd/jobs/list_resource_job.go

package cicdjobs

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"

	cicdclient "github.com/SAP/terraform-provider-btp-services/internal/cicd/client"
	"github.com/SAP/terraform-provider-btp-services/internal/shared"
)

var _ list.ListResourceWithConfigure = &jobListResource{}

type jobListResource struct {
	cli *cicdclient.CicdClientFacade
}

func NewJobListResource() list.ListResource {
	return &jobListResource{}
}

func (r *jobListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_cicd_job", req.ProviderTypeName)
}

func (r *jobListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *jobListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{
		MarkdownDescription: "This list resource discovers all jobs in the SAP BTP CI/CD service. " +
			"All filter attributes are optional and applied in-memory after fetching.",
		Attributes: map[string]listschema.Attribute{
			"pipeline": listschema.StringAttribute{
				MarkdownDescription: "Filter by pipeline type. One of: `cpi`, `cf-env`, `kyma-cnb`, `sap-ui5-abap-fes`.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"cpi",
						"cf-env",
						"kyma-cnb",
						"sap-ui5-abap-fes",
					),
				},
			},
			"repository_id": listschema.StringAttribute{
				MarkdownDescription: "Filter by repository ID. Only returns jobs that belong to this repository.",
				Optional:            true,
			},
		},
	}
}

// jobListConfigModel holds the user-supplied filter values from the list block.
type jobListConfigModel struct {
	Pipeline     types.String `tfsdk:"pipeline"`
	RepositoryID types.String `tfsdk:"repository_id"`
}

func (r *jobListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var config jobListConfigModel
	var diags diag.Diagnostics
	diags.Append(req.Config.Get(ctx, &config)...)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	jobs, err := r.cli.Jobs.List(ctx)
	if err != nil {
		diags.AddError(
			"API Error Listing CI/CD Jobs",
			fmt.Sprintf("Failed to list jobs: %s", err),
		)
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, job := range jobs {
			if !config.Pipeline.IsNull() && job.Pipeline != config.Pipeline.ValueString() {
				continue
			}
			if !config.RepositoryID.IsNull() && job.RepositoryID != config.RepositoryID.ValueString() {
				continue
			}

			result := req.NewListResult(ctx)
			result.Identity.SetAttribute(ctx, path.Root("id"), job.ID)

			if req.IncludeResource {
				model, err := jobResourceValueFrom(job, "", true)
				if err != nil {
					result.Diagnostics.AddError("Error Mapping Job Response", err.Error())
					push(result)
					return
				}
				result.Diagnostics.Append(result.Resource.Set(ctx, model)...)
			}

			if !push(result) {
				return
			}
		}
	}
}
