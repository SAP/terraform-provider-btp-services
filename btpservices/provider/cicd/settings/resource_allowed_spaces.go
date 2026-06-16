// btpservices/provider/cicd/settings/resource_allowed_spaces.go

package cicdsettings

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"

	cicdclient "github.com/SAP/terraform-provider-btp-services/internal/cicd/client"
	cicdmodels "github.com/SAP/terraform-provider-btp-services/internal/cicd/models"
	"github.com/SAP/terraform-provider-btp-services/internal/shared"
)

var _ resource.Resource = &allowedSpacesResource{}
var _ resource.ResourceWithConfigure = &allowedSpacesResource{}

func NewAllowedSpacesResource() resource.Resource {
	return &allowedSpacesResource{}
}

type allowedSpacesResource struct {
	cli *cicdclient.CicdClientFacade
}

func (r *allowedSpacesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_cicd_allowed_spaces", req.ProviderTypeName)
}

func (r *allowedSpacesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the list of allowed spaces in the SAP BTP CI/CD service.\n\n" +
			"This is a **bulk-replace** resource: the entire list is replaced on every apply. " +
			"Spaces not listed will lose access to CI/CD service instances.\n\n" +
			"~> **Warning:** Only one `btpservice_cicd_allowed_spaces` resource must exist per provider configuration. " +
			"Defining multiple instances will cause them to overwrite each other on every apply, " +
			"leading to unpredictable state and loss of access for spaces managed by the other instance.",
		Attributes: map[string]schema.Attribute{
			"allowed_spaces": schema.SetNestedAttribute{
				MarkdownDescription: "Set of Cloud Foundry spaces that are permitted to request CI/CD service instances.",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"space_guid": schema.StringAttribute{
							MarkdownDescription: "GUID of the Cloud Foundry space (UUID format).",
							Required:            true,
						},
						"comment": schema.StringAttribute{
							MarkdownDescription: "Human-readable note about why this space is allowed (max 255 characters).",
							Required:            true,
						},
					},
				},
			},
		},
	}
}

func (r *allowedSpacesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *allowedSpacesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan allowedSpacesModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.cli.AllowedSpaces.Set(ctx, plan.toRequest()); err != nil {
		resp.Diagnostics.AddError("Error Setting Allowed Spaces", err.Error())
		return
	}

	result, err := r.cli.AllowedSpaces.Get(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Allowed Spaces After Create", err.Error())
		return
	}

	state := allowedSpacesValueFrom(*result)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *allowedSpacesResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	result, err := r.cli.AllowedSpaces.Get(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Allowed Spaces", err.Error())
		return
	}

	state := allowedSpacesValueFrom(*result)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *allowedSpacesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan allowedSpacesModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.cli.AllowedSpaces.Set(ctx, plan.toRequest()); err != nil {
		resp.Diagnostics.AddError("Error Updating Allowed Spaces", err.Error())
		return
	}

	result, err := r.cli.AllowedSpaces.Get(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Allowed Spaces After Update", err.Error())
		return
	}

	state := allowedSpacesValueFrom(*result)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *allowedSpacesResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Clear the list to revoke all space access when the resource is destroyed.
	if err := r.cli.AllowedSpaces.Set(ctx, cicdmodels.AllowedSpaceListDTO{AllowedSpaces: []cicdmodels.AllowedSpace{}}); err != nil {
		resp.Diagnostics.AddError("Error Clearing Allowed Spaces", err.Error())
	}
}
