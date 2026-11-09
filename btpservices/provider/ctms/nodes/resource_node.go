package ctmsnodes

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/utils"
	ctmsmodels "github.com/SAP/terraform-provider-btp-services/internal/ctms/models"
)

var _ resource.ResourceWithConfigure = &nodeResource{}
var _ resource.ResourceWithImportState = &nodeResource{}
var _ resource.ResourceWithIdentity = &nodeResource{}
var _ resource.ResourceWithValidateConfig = &nodeResource{}

type nodeResource struct{ utils.Connection }

func NewNodeResource() resource.Resource {
	return &nodeResource{}
}

func (r *nodeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ctms_node"
}

func (r *nodeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a CTMS node. Requires Terraform 1.14+. Import using its decimal ID or an identity containing " +
			"endpoint, subaccount_id, and id. Omitted optional settings remain managed by CTMS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Decimal node ID assigned by CTMS.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Case-sensitive node name. Changing it replaces the node.",
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Node description. Omit to leave it managed by CTMS.",
			},
			"upload_allowed": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Allow file uploads to this node.",
			},
			"forward_mode": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Forwarding mode: AUTO (pre-import), POST_IMPORT, ON_SUCCESS, or MANUAL.",
				Validators:          []validator.String{stringvalidator.OneOf("AUTO", "POST_IMPORT", "ON_SUCCESS", "MANUAL")},
			},
			"notification_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Enable notifications. Requires an Alert Notification subscription and configuration in the subaccount.",
			},
			"import_disabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Disable imports, including scheduled and automatic imports; forwarding can still occur.",
			},
			"import_disabled_reason": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional reason for disabling imports.",
			},
			"virtual": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Virtual nodes cannot have deployment targets.",
			},
			"auto_import": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				MarkdownDescription: "Automatically import eligible requests. Enabling this can immediately import queued requests and " +
					"deactivate existing schedules. It can coexist with import_disabled; imports resume when that flag is " +
					"cleared.",
			},
			"controlled_by": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Application controlling the node. Read-only until the API reset contract is confirmed.",
			},
			"job_id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Import scheduler job ID, if present.",
			},
			"targets": schema.ListNestedAttribute{
				Optional: true,
				Computed: true,
				MarkdownDescription: "Deployment targets in API order. Omit to leave targets managed by CTMS; an explicit empty list " +
					"requests clearing them. Virtual nodes cannot have targets.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"content_type": schema.StringAttribute{
							Required:            true,
							Validators:          []validator.String{stringvalidator.OneOf("MTA", "XSCDU", "APP", "BTP_ABAP")},
							MarkdownDescription: "Content type.",
						},
						"destination_name": schema.StringAttribute{
							Required:            true,
							Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
							MarkdownDescription: "Name of an existing transport destination.",
						},
						"import_options": schema.SingleNestedAttribute{
							Optional:            true,
							Computed:            true,
							MarkdownDescription: "MTA deployment options. The destination must address Cloud Foundry.",
							Attributes: map[string]schema.Attribute{
								"strategy": schema.StringAttribute{
									Optional:            true,
									Validators:          []validator.String{stringvalidator.OneOf("default", "blue-green")},
									Computed:            true,
									MarkdownDescription: "Deployment strategy: default or blue-green. Only for MTA deployment to Cloud Foundry.",
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *nodeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(r.Connection.Configure(req.ProviderData)...)
}

func (r *nodeResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = utils.IdentitySchema()
}

func (r *nodeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	r.Import(ctx, req, resp)
}

func (r *nodeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan nodeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	targets, diags := targetsRequest(ctx, plan.Targets)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	request := ctmsmodels.NodeCreateRequest{
		Name:                 plan.Name.ValueString(),
		Targets:              targets,
		Description:          utils.StringPointer[string](plan.Description),
		UploadAllowed:        utils.BoolPointer(plan.UploadAllowed),
		ForwardMode:          utils.StringPointer[ctmsmodels.ForwardMode](plan.ForwardMode),
		NotificationEnabled:  utils.BoolPointer(plan.NotificationEnabled),
		ImportDisabled:       utils.BoolPointer(plan.ImportDisabled),
		ImportDisabledReason: utils.StringPointer[string](plan.ImportDisabledReason),
		Virtual:              utils.BoolPointer(plan.Virtual),
		AutoImport:           utils.BoolPointer(plan.AutoImport),
	}
	result, err := r.Client.Nodes.Create(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating CTMS Node", utils.CreateError("node", err))
		return
	}
	model, d := nodeValueFrom(ctx, *result)
	resp.Diagnostics.Append(d...)
	// Preserve the assigned identity even if the response is incomplete.
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, r.Identity(model.ID))...)
}

func (r *nodeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state nodeModel
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
	result, err := r.Client.Nodes.Get(ctx, id)
	if ctmsmodels.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error Reading CTMS Node", err.Error())
		return
	}
	model, d := nodeValueFrom(ctx, *result)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, r.Identity(model.ID))...)
}

func (r *nodeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, config, state nodeModel
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
	request := ctmsmodels.NodeUpdateRequest{}
	if !config.Description.IsNull() && !plan.Description.Equal(state.Description) {
		request.Description = utils.StringPointer[string](plan.Description)
	}
	if !config.UploadAllowed.IsNull() && !plan.UploadAllowed.Equal(state.UploadAllowed) {
		request.UploadAllowed = utils.BoolPointer(plan.UploadAllowed)
	}
	if !config.ForwardMode.IsNull() && !plan.ForwardMode.Equal(state.ForwardMode) {
		request.ForwardMode = utils.StringPointer[ctmsmodels.ForwardMode](plan.ForwardMode)
	}
	if !config.NotificationEnabled.IsNull() && !plan.NotificationEnabled.Equal(state.NotificationEnabled) {
		request.NotificationEnabled = utils.BoolPointer(plan.NotificationEnabled)
	}
	if !config.ImportDisabled.IsNull() && !plan.ImportDisabled.Equal(state.ImportDisabled) {
		request.ImportDisabled = utils.BoolPointer(plan.ImportDisabled)
	}
	if !config.ImportDisabledReason.IsNull() && !plan.ImportDisabledReason.Equal(state.ImportDisabledReason) {
		request.ImportDisabledReason = utils.StringPointer[string](plan.ImportDisabledReason)
	}
	if !config.Virtual.IsNull() && !plan.Virtual.Equal(state.Virtual) {
		request.Virtual = utils.BoolPointer(plan.Virtual)
	}
	if !config.AutoImport.IsNull() && !plan.AutoImport.Equal(state.AutoImport) {
		request.AutoImport = utils.BoolPointer(plan.AutoImport)
	}
	if !config.Targets.IsNull() && !plan.Targets.Equal(state.Targets) {
		targets, d := targetsRequest(ctx, plan.Targets)
		resp.Diagnostics.Append(d...)
		request.Targets = targets
	}
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.Client.Nodes.Update(ctx, id, request)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating CTMS Node", err.Error())
		return
	}
	model, d := nodeValueFrom(ctx, *result)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, r.Identity(model.ID))...)
}

func (r *nodeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state nodeModel
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
	if err := r.Client.Nodes.Delete(ctx, id); err != nil && !ctmsmodels.IsNotFound(err) {
		resp.Diagnostics.AddError("Error Deleting CTMS Node", err.Error()+" Remove or rewire routes referencing this node before deleting it.")
	}
}

func (r *nodeResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config nodeModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() || config.Targets.IsNull() || config.Targets.IsUnknown() {
		return
	}
	if config.Virtual.ValueBool() && len(config.Targets.Elements()) > 0 {
		resp.Diagnostics.AddAttributeError(path.Root("targets"), "Virtual Node Has Targets", "A virtual node cannot have deployment targets.")
	}
	for i, value := range config.Targets.Elements() {
		if value.IsNull() || value.IsUnknown() {
			continue
		}
		target, ok := value.(types.Object)
		if !ok {
			continue
		}
		contentType, ok := target.Attributes()["content_type"].(types.String)
		if !ok || contentType.IsNull() || contentType.IsUnknown() {
			continue
		}
		options, ok := target.Attributes()["import_options"].(types.Object)
		if !ok || options.IsNull() || options.IsUnknown() {
			continue
		}
		strategy := options.Attributes()["strategy"]
		if strategy != nil && !strategy.IsNull() && !strategy.IsUnknown() && contentType.ValueString() != "MTA" {
			resp.Diagnostics.AddAttributeError(
				path.Root("targets").AtListIndex(i).AtName("import_options").AtName("strategy"),
				"Invalid Deployment Strategy",
				"A deployment strategy requires content_type = MTA and a destination addressing Cloud Foundry.",
			)
		}
	}
}
