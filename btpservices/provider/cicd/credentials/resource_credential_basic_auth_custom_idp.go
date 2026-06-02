// btpservices/provider/cicd/credentials/resource_credential_basic_auth_custom_idp.go

package cicdcredentials

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	cicdclient "github.com/SAP/terraform-provider-btp-services/internal/cicd/client"
	cicdmodels "github.com/SAP/terraform-provider-btp-services/internal/cicd/models"
	"github.com/SAP/terraform-provider-btp-services/internal/shared"
)

var _ resource.Resource = &basicAuthCIdPResource{}
var _ resource.ResourceWithConfigure = &basicAuthCIdPResource{}

func NewBasicAuthCIdPResource() resource.Resource {
	return &basicAuthCIdPResource{}
}

type basicAuthCIdPResource struct {
	cli *cicdclient.CicdClientFacade
}

func (r *basicAuthCIdPResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_cicd_credential_basic_auth_custom_idp", req.ProviderTypeName)
}

func (r *basicAuthCIdPResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Basic Authentication with Custom Identity Provider credential in the SAP BTP CI/CD service.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the resource (assigned by the API).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Credential name. Must contain only lowercase letters, numbers, and hyphens.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Optional human-readable description.",
				Optional:            true,
				Computed:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username for basic authentication.",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for basic authentication. Not returned by the API on reads — stored only in Terraform state.",
				Required:            true,
				Sensitive:           true,
			},
			"origin": schema.StringAttribute{
				MarkdownDescription: "The custom identity provider's origin key (e.g. \"custom-platform\").",
				Required:            true,
			},
		},
	}
}

func (r *basicAuthCIdPResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *basicAuthCIdPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan basicAuthCIdPResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.cli.Credentials.Create(ctx, plan.toCreateRequest()); err != nil {
		resp.Diagnostics.AddError("Error Creating Credential", err.Error())
		return
	}

	result, err := r.cli.Credentials.Get(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Credential After Create", err.Error())
		return
	}

	state := basicAuthCIdPResourceValueFrom(*result)
	state.Password = plan.Password
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *basicAuthCIdPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state basicAuthCIdPResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.cli.Credentials.Get(ctx, state.ID.ValueString())
	if err != nil {
		if cicdmodels.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Credential", err.Error())
		return
	}

	updated := basicAuthCIdPResourceValueFrom(*result)
	updated.Password = state.Password
	resp.Diagnostics.Append(resp.State.Set(ctx, updated)...)
}

func (r *basicAuthCIdPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan basicAuthCIdPResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state basicAuthCIdPResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.cli.Credentials.Patch(ctx, state.ID.ValueString(), plan.toPatchRequest()); err != nil {
		resp.Diagnostics.AddError("Error Updating Credential", err.Error())
		return
	}

	result, err := r.cli.Credentials.Get(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Credential After Update", err.Error())
		return
	}

	updated := basicAuthCIdPResourceValueFrom(*result)
	updated.Password = plan.Password
	resp.Diagnostics.Append(resp.State.Set(ctx, updated)...)
}

func (r *basicAuthCIdPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state basicAuthCIdPResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.cli.Credentials.Delete(ctx, state.ID.ValueString()); err != nil {
		if !cicdmodels.IsNotFound(err) {
			resp.Diagnostics.AddError("Error Deleting Credential", err.Error())
		}
	}
}
