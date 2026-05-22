// btpservices/provider/cicd/jobs/resource_build_trigger.go

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

	cicdclient "github.com/SAP/terraform-provider-sap-btp-services/internal/cicd/client"
	cicdmodels "github.com/SAP/terraform-provider-sap-btp-services/internal/cicd/models"
	"github.com/SAP/terraform-provider-sap-btp-services/internal/shared"
)

var _ resource.Resource = &buildTriggerResource{}
var _ resource.ResourceWithConfigure = &buildTriggerResource{}
var _ resource.ResourceWithImportState = &buildTriggerResource{}
var _ resource.ResourceWithIdentity = &buildTriggerResource{}

func NewBuildTriggerResource() resource.Resource {
	return &buildTriggerResource{}
}

type buildTriggerResource struct {
	cli *cicdclient.CicdClientFacade
}

func (r *buildTriggerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_cicd_build_trigger", req.ProviderTypeName)
}

func (r *buildTriggerResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
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

func (r *buildTriggerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a build trigger for a CI/CD job in the SAP BTP CI/CD service. Currently only `timer` type triggers are supported.",
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
				MarkdownDescription: "Trigger type. Currently the only supported value is `timer`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("timer"),
				},
			},
			"timer": schema.SingleNestedAttribute{
				MarkdownDescription: "Timer schedule configuration. Required when `type` is `TIMER`.",
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

func (r *buildTriggerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *buildTriggerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan buildTriggerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.cli.Jobs.CreateTrigger(ctx, plan.Job.ValueString(), plan.toCreateRequest())
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Build Trigger", err.Error())
		return
	}

	state := buildTriggerResourceValueFrom(plan.Job.ValueString(), *result)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, buildTriggerIdentityModel{
		Job: plan.Job,
		ID:  types.StringValue(result.ID),
	})...)
}

func (r *buildTriggerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state buildTriggerResourceModel
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
		resp.Diagnostics.AddError("Error Reading Build Trigger", err.Error())
		return
	}

	updated := buildTriggerResourceValueFrom(state.Job.ValueString(), *result)
	resp.Diagnostics.Append(resp.State.Set(ctx, updated)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, buildTriggerIdentityModel{
		Job: state.Job,
		ID:  state.ID,
	})...)
}

func (r *buildTriggerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan buildTriggerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state buildTriggerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.cli.Jobs.UpdateTrigger(ctx, state.Job.ValueString(), state.ID.ValueString(), plan.toUpdateRequest()); err != nil {
		resp.Diagnostics.AddError("Error Updating Build Trigger", err.Error())
		return
	}

	result, err := r.cli.Jobs.GetTrigger(ctx, state.Job.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Build Trigger After Update", err.Error())
		return
	}

	updated := buildTriggerResourceValueFrom(state.Job.ValueString(), *result)
	resp.Diagnostics.Append(resp.State.Set(ctx, updated)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, buildTriggerIdentityModel{
		Job: state.Job,
		ID:  state.ID,
	})...)
}

func (r *buildTriggerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state buildTriggerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.cli.Jobs.DeleteTrigger(ctx, state.Job.ValueString(), state.ID.ValueString()); err != nil {
		if !cicdmodels.IsNotFound(err) {
			resp.Diagnostics.AddError("Error Deleting Build Trigger", err.Error())
		}
	}
}

// ImportState expects an import ID in "job_ref/trigger_id" format.
func (r *buildTriggerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected format: <job_id>/<trigger_id>, got: %q", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("job"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, buildTriggerIdentityModel{
		Job: types.StringValue(parts[0]),
		ID:  types.StringValue(parts[1]),
	})...)
}
