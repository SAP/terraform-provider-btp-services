// btpservices/provider/cicd/repositories/resource_repository.go

package cicdrepositories

import (
	"context"
	"fmt"

	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	cicdclient "github.com/SAP/terraform-provider-btp-services/internal/cicd/client"
	cicdmodels "github.com/SAP/terraform-provider-btp-services/internal/cicd/models"
	"github.com/SAP/terraform-provider-btp-services/internal/shared"
)

var _ resource.Resource = &repositoryResource{}
var _ resource.ResourceWithConfigure = &repositoryResource{}
var _ resource.ResourceWithImportState = &repositoryResource{}
var _ resource.ResourceWithValidateConfig = &repositoryResource{}
var _ resource.ResourceWithIdentity = &repositoryResource{}

func NewRepositoryResource() resource.Resource {
	return &repositoryResource{}
}

type repositoryResource struct {
	cli *cicdclient.CicdClientFacade
}

func (r *repositoryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_cicd_repository", req.ProviderTypeName)
}

func (r *repositoryResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{
				RequiredForImport: true,
			},
		},
	}
}

func (r *repositoryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a repository resource in the SAP BTP CI/CD service.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the repository (assigned by the API on creation).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the repository. Allowed characters: `[a-zA-Z0-9_-]`, max 64 characters.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`),
						"must match [a-zA-Z0-9_-]{1,64}",
					),
				},
			},
			"clone_url": schema.StringAttribute{
				MarkdownDescription: "URL used to clone the repository (e.g. `https://github.com/example/repo`). Max 255 characters.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^https?://`),
						"must be a valid URL starting with http:// or https://",
					),
					stringvalidator.LengthAtMost(255),
				},
			},
			"clone_credential_id": schema.StringAttribute{
				MarkdownDescription: "ID of the credential used to authenticate against the Git server when cloning. Max 63 characters.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(63),
				},
			},
			"cloud_connector_credential_id": schema.StringAttribute{
				MarkdownDescription: "ID of the credential containing the Cloud Connector Location ID used for cloning. Max 63 characters.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(63),
				},
			},
			"event_receiver": schema.SingleNestedAttribute{
				MarkdownDescription: "Event receiver configuration for triggering builds via webhooks from your source code management system.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"active": schema.BoolAttribute{
						MarkdownDescription: "Whether event processing is enabled for this repository.",
						Required:            true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"scm_type": schema.StringAttribute{
						MarkdownDescription: "Source code manager type. Allowed values: `GITHUB`, `BITBUCKET_CLOUD`, `BITBUCKET`, `GITLAB`, `AZUREREPOS`.",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("GITHUB", "BITBUCKET_CLOUD", "BITBUCKET", "GITLAB", "AZUREREPOS"),
						},
					},
					"webhook_id": schema.StringAttribute{
						MarkdownDescription: "Immutable technical ID of the webhook, assigned by the API on creation.",
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},

					// webhook_token_credential_id is required for all SCM types except BITBUCKET_CLOUD.
					// Enforced by webhookTokenRequiredValidator on the event_receiver block.
					"webhook_token_credential_id": schema.StringAttribute{
						MarkdownDescription: "The ID for the webhook secret of this event receiver. Required for all SCM types except `BITBUCKET_CLOUD`. Max 63 characters.",
						Optional:            true,
						Computed:            true,
						Validators: []validator.String{
							stringvalidator.LengthAtMost(63),
						},
					},
				},
			},
		},
	}
}

func (r *repositoryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *repositoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan repositoryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.cli.Repositories.Create(ctx, plan.toCreateRequest()); err != nil {
		resp.Diagnostics.AddError("Error Creating Repository", err.Error())
		return
	}

	result, err := r.cli.Repositories.Get(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Repository After Create", err.Error())
		return
	}

	state := repositoryResourceValueFrom(*result)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, repositoryIdentityModel{ID: types.StringValue(result.ID)})...)
}

func (r *repositoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state repositoryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.cli.Repositories.Get(ctx, state.ID.ValueString())
	if err != nil {
		if cicdmodels.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Repository", err.Error())
		return
	}

	updated := repositoryResourceValueFrom(*result)
	resp.Diagnostics.Append(resp.State.Set(ctx, updated)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, repositoryIdentityModel{ID: types.StringValue(result.ID)})...)
}

func (r *repositoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan repositoryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state repositoryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.cli.Repositories.Update(ctx, plan.toUpdateRequest(state)); err != nil {
		resp.Diagnostics.AddError("Error Updating Repository", err.Error())
		return
	}

	result, err := r.cli.Repositories.Get(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Repository After Update", err.Error())
		return
	}

	updated := repositoryResourceValueFrom(*result)
	resp.Diagnostics.Append(resp.State.Set(ctx, updated)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, repositoryIdentityModel{ID: types.StringValue(result.ID)})...)
}

func (r *repositoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state repositoryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.cli.Repositories.Delete(ctx, state.ID.ValueString()); err != nil {
		if !cicdmodels.IsNotFound(err) {
			resp.Diagnostics.AddError("Error Deleting Repository", err.Error())
		}
	}
}

func (r *repositoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("id"), req, resp)
}

func (r *repositoryResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data repositoryResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() || data.EventReceiver == nil {
		return
	}
	scm := data.EventReceiver.SCMType
	token := data.EventReceiver.WebhookTokenCredentialID
	if scm.IsUnknown() || token.IsUnknown() {
		return
	}
	if scm.ValueString() != "BITBUCKET_CLOUD" && (token.IsNull() || token.ValueString() == "") {
		resp.Diagnostics.AddAttributeError(
			path.Root("event_receiver").AtName("webhook_token_credential_id"),
			"Missing Required Attribute",
			"webhook_token_credential_id is required when scm_type is not BITBUCKET_CLOUD.",
		)
	}
}
