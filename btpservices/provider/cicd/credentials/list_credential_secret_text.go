package cicdcredentials

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	cicdclient "github.com/SAP/terraform-provider-sap-btp-services/internal/cicd/client"
	"github.com/SAP/terraform-provider-sap-btp-services/internal/shared"
)

var _ list.ListResource = &secretTextListResource{}
var _ list.ListResourceWithConfigure = &secretTextListResource{}

func NewSecretTextListResource() list.ListResource {
	return &secretTextListResource{}
}

type secretTextListResource struct {
	cli *cicdclient.CicdClientFacade
}

func (r *secretTextListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_cicd_credential_secret_text", req.ProviderTypeName)
}

func (r *secretTextListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{
		MarkdownDescription: "This list resource discovers all Secret Text credentials in the SAP BTP CI/CD service.",
	}
}

func (r *secretTextListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *secretTextListResource) List(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) {
	creds, err := r.cli.Credentials.List(ctx)
	if err != nil {
		resp.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Error Listing Credentials", err.Error()),
		})
		return
	}

	resp.Results = func(yield func(list.ListResult) bool) {
		for _, cred := range creds {
			if cred.SecretText == nil {
				continue
			}
			result := req.NewListResult(ctx)
			result.DisplayName = cred.Name
			result.Diagnostics.Append(result.Identity.Set(ctx, credentialIdentityModel{
				ID: types.StringValue(cred.ID),
			})...)
			if req.IncludeResource {
				model := secretTextResourceValueFrom(cred)
				result.Diagnostics.Append(result.Resource.Set(ctx, &model)...)
			}
			if !yield(result) {
				return
			}
		}
	}
}
