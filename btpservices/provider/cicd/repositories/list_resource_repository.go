// btpservices/provider/cicd/repositories/list_resource_repository.go

package cicdrepositories

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	cicdclient "github.com/SAP/terraform-provider-sap-btp-services/internal/cicd/client"
	"github.com/SAP/terraform-provider-sap-btp-services/internal/shared"
)

var _ list.ListResourceWithConfigure = &repositoryListResource{}

type repositoryListResource struct {
	cli *cicdclient.CicdClientFacade
}

func NewRepositoryListResource() list.ListResource {
	return &repositoryListResource{}
}

func (r *repositoryListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_cicd_repository", req.ProviderTypeName)
}

func (r *repositoryListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *repositoryListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{
		MarkdownDescription: "This list resource discovers all repositories in the SAP BTP CI/CD service.",
	}
}

func (r *repositoryListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	repos, err := r.cli.Repositories.List(ctx)
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError(
			"API Error Listing CI/CD Repositories",
			fmt.Sprintf("Failed to list repositories: %s", err),
		)
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, repo := range repos {
			result := req.NewListResult(ctx)
			result.Identity.SetAttribute(ctx, path.Root("id"), repo.ID)

			if req.IncludeResource {
				model := repositoryResourceValueFrom(repo)
				result.Diagnostics.Append(result.Resource.Set(ctx, model)...)
			}

			if !push(result) {
				return
			}
		}
	}
}
