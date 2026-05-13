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

var _ list.ListResource = &basicAuthCIdPListResource{}
var _ list.ListResourceWithConfigure = &basicAuthCIdPListResource{}

func NewBasicAuthCIdPListResource() list.ListResource {
	return &basicAuthCIdPListResource{}
}

type basicAuthCIdPListResource struct {
	cli *cicdclient.CicdClientFacade
}

func (r *basicAuthCIdPListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_cicd_credential_basic_auth_custom_idp", req.ProviderTypeName)
}

func (r *basicAuthCIdPListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{
		MarkdownDescription: "This list resource discovers all Basic Authentication credentials for custom Identity Providers in the SAP BTP CI/CD service.",
	}
}

func (r *basicAuthCIdPListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *basicAuthCIdPListResource) List(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) {
	creds, err := r.cli.Credentials.List(ctx)
	if err != nil {
		resp.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Error Listing Credentials", err.Error()),
		})
		return
	}

	resp.Results = func(yield func(list.ListResult) bool) {
		for _, cred := range creds {
			if cred.BasicForCustomIdP == nil {
				continue
			}
			result := req.NewListResult(ctx)
			result.DisplayName = cred.Name
			result.Diagnostics.Append(result.Identity.Set(ctx, credentialIdentityModel{
				ID: types.StringValue(cred.ID),
			})...)
			if req.IncludeResource {
				model := basicAuthCIdPResourceValueFrom(cred)
				result.Diagnostics.Append(result.Resource.Set(ctx, &model)...)
			}
			if !yield(result) {
				return
			}
		}
	}
}
