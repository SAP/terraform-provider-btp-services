package ctmsroutes

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/utils"
	ctmsmodels "github.com/SAP/terraform-provider-btp-services/internal/ctms/models"
)

var _ resource.ResourceWithConfigure = &routeResource{}
var _ resource.ResourceWithImportState = &routeResource{}
var _ resource.ResourceWithIdentity = &routeResource{}

type routeResource struct{ utils.Connection }

func NewRouteResource() resource.Resource {
	return &routeResource{}
}

func (r *routeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ctms_route"
}

func (r *routeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a CTMS route. Requires Terraform 1.14+. Import using its decimal ID or an identity " +
			"containing endpoint, subaccount_id, and id. Omitted optional settings remain managed by CTMS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Decimal route ID assigned by CTMS.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Route name. Supports in-place updates.",
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Route description. Omit to leave it managed by CTMS.",
			},
			"source_node_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Decimal source node ID. A source can have multiple outgoing routes.",
				Validators:          []validator.String{utils.IDValidator()},
			},
			"target_node_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Decimal target node ID. A node can be the target of only one route.",
				Validators:          []validator.String{utils.IDValidator()},
			},
		},
	}
}

func (r *routeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(r.Connection.Configure(req.ProviderData)...)
}

func (r *routeResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = utils.IdentitySchema()
}

func (r *routeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	r.Import(ctx, req, resp)
}

func (r *routeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan routeModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	source, err := utils.ParseID(plan.SourceNodeID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid CTMS Source ID", err.Error())
		return
	}

	target, err := utils.ParseID(plan.TargetNodeID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid CTMS Target ID", err.Error())
		return
	}

	request := ctmsmodels.RouteCreateRequest{
		Name:         plan.Name.ValueString(),
		Description:  utils.StringPointer[string](plan.Description),
		SourceNodeId: source,
		TargetNodeId: target,
	}

	result, err := r.Client.Routes.Create(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating CTMS Route", utils.CreateError("route", err))
		return
	}

	model, d := routeValueFrom(ctx, *result)

	resp.Diagnostics.Append(d...)
	// Preserve the assigned identity even if the response is incomplete.
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, r.Identity(model.ID))...)
}

func (r *routeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state routeModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(r.CheckIdentity(ctx, req.Identity, state.ID)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := utils.ParseID(state.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid CTMS ID", err.Error())
		return
	}

	result, err := r.Client.Routes.Get(ctx, id)
	if ctmsmodels.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Error Reading CTMS Route", err.Error())
		return
	}

	model, d := routeValueFrom(ctx, *result)

	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, r.Identity(model.ID))...)
}

func (r *routeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, config, state routeModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(r.CheckIdentity(ctx, req.Identity, state.ID)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := utils.ParseID(state.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid CTMS ID", err.Error())
		return
	}

	request := ctmsmodels.RoutePatchRequest{}
	if !config.Name.IsNull() && !plan.Name.Equal(state.Name) {
		request.Name = utils.StringPointer[string](plan.Name)
	}

	if !config.Description.IsNull() && !plan.Description.Equal(state.Description) {
		request.Description = utils.StringPointer[string](plan.Description)
	}

	if !config.SourceNodeID.IsNull() && !plan.SourceNodeID.Equal(state.SourceNodeID) {
		value, e := utils.ParseID(plan.SourceNodeID)
		if e != nil {
			resp.Diagnostics.AddError("Invalid CTMS Node ID", e.Error())
			return
		}
		request.SourceNodeId = &value
	}

	if !config.TargetNodeID.IsNull() && !plan.TargetNodeID.Equal(state.TargetNodeID) {
		value, e := utils.ParseID(plan.TargetNodeID)
		if e != nil {
			resp.Diagnostics.AddError("Invalid CTMS Node ID", e.Error())
			return
		}
		request.TargetNodeId = &value
	}

	result, err := r.Client.Routes.Update(ctx, id, request)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating CTMS Route", err.Error())
		return
	}

	model, d := routeValueFrom(ctx, *result)

	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, r.Identity(model.ID))...)
}

func (r *routeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state routeModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(r.CheckIdentity(ctx, req.Identity, state.ID)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := utils.ParseID(state.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid CTMS ID", err.Error())
		return
	}

	if err := r.Client.Routes.Delete(ctx, id); err != nil && !ctmsmodels.IsNotFound(err) {
		resp.Diagnostics.AddError("Error Deleting CTMS Route", err.Error()+"")
	}
}
