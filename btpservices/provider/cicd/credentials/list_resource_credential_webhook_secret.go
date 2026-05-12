// btpservices/provider/cicd/credentials/list_resource_credential_webhook_secret.go

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

var _ list.ListResourceWithConfigure = &webhookSecretListResource{}

type webhookSecretListResource struct {
	cli *cicdclient.CicdClientFacade
}

func NewWebhookSecretListResource() list.ListResource {
	return &webhookSecretListResource{}
}

func (r *webhookSecretListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_cicd_credential_webhook_secret", req.ProviderTypeName)
}

func (r *webhookSecretListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *webhookSecretListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{
		MarkdownDescription: "This list resource discovers all Webhook Secret credentials in the SAP BTP CI/CD service.",
	}
}

func (r *webhookSecretListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	creds, err := r.cli.Credentials.List(ctx)
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError(
			"API Error Listing CI/CD Credentials (Webhook Secret)",
			fmt.Sprintf("Failed to list credentials: %s", err),
		)
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, cred := range creds {
			if cred.WebhookToken == nil {
				continue
			}

			result := req.NewListResult(ctx)
			result.DisplayName = cred.Name
			result.Diagnostics.Append(result.Identity.Set(ctx, credentialIdentityModel{
				ID: types.StringValue(cred.ID),
			})...)

			if req.IncludeResource {
				model := webhookSecretResourceValueFrom(cred)
				result.Diagnostics.Append(result.Resource.Set(ctx, model)...)
			}

			if !push(result) {
				return
			}
		}
	}
}
