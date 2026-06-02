// btpservices/provider/cicd/jobs/resource_trigger.go

package cicdjobs

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	cicdclient "github.com/SAP/terraform-provider-btp-services/internal/cicd/client"
	cicdmodels "github.com/SAP/terraform-provider-btp-services/internal/cicd/models"
	"github.com/SAP/terraform-provider-btp-services/internal/shared"
)

var _ resource.Resource = &triggerResource{}
var _ resource.ResourceWithConfigure = &triggerResource{}
var _ resource.ResourceWithImportState = &triggerResource{}
var _ resource.ResourceWithIdentity = &triggerResource{}
var _ resource.ResourceWithValidateConfig = &triggerResource{}

func NewTriggerResource() resource.Resource {
	return &triggerResource{}
}

type triggerResource struct {
	cli *cicdclient.CicdClientFacade
}

func (r *triggerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_cicd_trigger", req.ProviderTypeName)
}

func (r *triggerResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"job": identityschema.StringAttribute{
				RequiredForImport: true,
			},
			"id": identityschema.StringAttribute{
				RequiredForImport: true,
			},
		},
	}
}

func (r *triggerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a trigger for a CI/CD job in the SAP BTP CI/CD service. Currently only `timer` type triggers are supported.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the trigger (assigned by the API on creation).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"job": schema.StringAttribute{
				MarkdownDescription: "Name or ID of the CI/CD job this trigger belongs to. Changing this forces recreation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Trigger type. Currently the only supported value is `timer`. Changing this forces recreation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("timer"),
				},
			},
			"timer": schema.SingleNestedAttribute{
				MarkdownDescription: "Timer schedule configuration. Required when `type` is `timer`.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"branch": schema.StringAttribute{
						MarkdownDescription: "Branch to build on the timer schedule.",
						Optional:            true,
						Computed:            true,
					},
					"cron": schema.StringAttribute{
						MarkdownDescription: "Cron expression defining the timer schedule (e.g. `0 9 * * 1-5` for weekdays at 09:00).",
						Optional:            true,
						Computed:            true,
					},
				},
			},
		},
	}
}

func (r *triggerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ValidateConfig enforces that the timer block is set when type = "timer".
func (r *triggerResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data triggerResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() || data.Type.IsUnknown() {
		return
	}
	if data.Type.ValueString() == "timer" && data.Timer == nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("timer"),
			"Missing Required Attribute",
			`timer block is required when type is "timer".`,
		)
	}
}

func (r *triggerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan triggerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.cli.Jobs.CreateTrigger(ctx, plan.Job.ValueString(), plan.toCreateRequest())
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Trigger", err.Error())
		return
	}

	state := triggerResourceValueFrom(plan.Job.ValueString(), *result)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, triggerIdentityModel{
		Job: plan.Job,
		ID:  types.StringValue(result.ID),
	})...)
}

func (r *triggerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state triggerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.cli.Jobs.GetTrigger(ctx, state.Job.ValueString(), state.ID.ValueString())
	if err != nil {
		if cicdmodels.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Trigger", err.Error())
		return
	}

	var identity triggerIdentityModel
	diags := req.Identity.Get(ctx, &identity)
	if diags.HasError() {
		identity = triggerIdentityModel{Job: state.Job, ID: state.ID}
	}

	updated := triggerResourceValueFrom(state.Job.ValueString(), *result)
	resp.Diagnostics.Append(resp.State.Set(ctx, updated)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identity)...)
}

func (r *triggerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan triggerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state triggerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.cli.Jobs.UpdateTrigger(ctx, state.Job.ValueString(), state.ID.ValueString(), plan.toUpdateRequest()); err != nil {
		resp.Diagnostics.AddError("Error Updating Trigger", err.Error())
		return
	}

	result, err := r.cli.Jobs.GetTrigger(ctx, state.Job.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Trigger After Update", err.Error())
		return
	}

	updated := triggerResourceValueFrom(state.Job.ValueString(), *result)
	resp.Diagnostics.Append(resp.State.Set(ctx, updated)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, triggerIdentityModel{
		Job: state.Job,
		ID:  state.ID,
	})...)
}

func (r *triggerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state triggerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.cli.Jobs.DeleteTrigger(ctx, state.Job.ValueString(), state.ID.ValueString()); err != nil {
		if !cicdmodels.IsNotFound(err) {
			resp.Diagnostics.AddError("Error Deleting Trigger", err.Error())
		}
	}
}

// ImportState accepts either "job_ref,trigger_id" (string-based import) or an
// identity-based import (Terraform 1.12+) where req.Identity carries job and id.
func (r *triggerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "" {
		parts := strings.SplitN(req.ID, ",", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			resp.Diagnostics.AddError(
				"Invalid Import ID",
				fmt.Sprintf("Expected format: <job_id>,<trigger_id>, got: %q", req.ID),
			)
			return
		}
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("job"), parts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
		resp.Diagnostics.Append(resp.Identity.Set(ctx, triggerIdentityModel{
			Job: types.StringValue(parts[0]),
			ID:  types.StringValue(parts[1]),
		})...)
		return
	}
	// Identity-based import (Terraform 1.12+): resp.Identity is pre-populated from req.Identity.
	var identity triggerIdentityModel
	diags := resp.Identity.Get(ctx, &identity)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("job"), identity.Job)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), identity.ID)...)
}
