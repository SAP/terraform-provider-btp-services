// btpservices/provider/cicd/repositories/types.go

package cicdrepositories

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	cicdmodels "github.com/SAP/terraform-provider-btp-services/internal/cicd/models"
)

// repositoryIdentityModel holds the stable identity of a repository resource.
type repositoryIdentityModel struct {
	ID types.String `tfsdk:"id"`
}

// eventReceiverModel is the Terraform state model for the event_receiver nested block.
// webhookId is computed — the API assigns it and never changes it.
type eventReceiverModel struct {
	Active                   types.Bool   `tfsdk:"active"`
	SCMType                  types.String `tfsdk:"scm_type"`
	WebhookID                types.String `tfsdk:"webhook_id"`
	WebhookTokenCredentialID types.String `tfsdk:"webhook_token_credential_id"`
}

// repositoryResourceModel is the Terraform state model for the repository resource.
type repositoryResourceModel struct {
	ID                         types.String        `tfsdk:"id"`
	Name                       types.String        `tfsdk:"name"`
	CloneURL                   types.String        `tfsdk:"clone_url"`
	CloneCredentialID          types.String        `tfsdk:"clone_credential_id"`
	CloudConnectorCredentialID types.String        `tfsdk:"cloud_connector_credential_id"`
	EventReceiver              *eventReceiverModel `tfsdk:"event_receiver"`
}

// optionalStringPtr converts an optional API string pointer to a Terraform types.String.
// A nil pointer maps to types.StringNull(); a non-nil pointer maps to types.StringValue().
func optionalStringPtr(s *string) types.String {
	if s == nil {
		return types.StringNull()
	}
	return types.StringValue(*s)
}

func repositoryResourceValueFrom(v cicdmodels.Repository) repositoryResourceModel {
	m := repositoryResourceModel{
		ID:                         types.StringValue(v.ID),
		Name:                       types.StringValue(v.Name),
		CloneURL:                   types.StringValue(v.CloneURL),
		CloneCredentialID:          optionalStringPtr(v.CloneCredentialID),
		CloudConnectorCredentialID: optionalStringPtr(v.CloudConnectorCredentialID),
	}
	if v.EventReceiver != nil {
		m.EventReceiver = &eventReceiverModel{
			Active:                   types.BoolValue(v.EventReceiver.Active),
			SCMType:                  types.StringValue(v.EventReceiver.SCMType),
			WebhookID:                types.StringValue(v.EventReceiver.WebhookID),
			WebhookTokenCredentialID: types.StringValue(v.EventReceiver.WebhookTokenCredentialID),
		}
	}
	return m
}

func (m repositoryResourceModel) toCreateRequest() cicdmodels.CreateRepositoryRequest {
	req := cicdmodels.CreateRepositoryRequest{
		Name:     m.Name.ValueString(),
		CloneURL: m.CloneURL.ValueString(),
	}
	if !m.CloneCredentialID.IsNull() && !m.CloneCredentialID.IsUnknown() {
		v := m.CloneCredentialID.ValueString()
		req.CloneCredentialID = &v
	}
	if !m.CloudConnectorCredentialID.IsNull() && !m.CloudConnectorCredentialID.IsUnknown() {
		v := m.CloudConnectorCredentialID.ValueString()
		req.CloudConnectorCredentialID = &v
	}
	if m.EventReceiver != nil {
		req.EventReceiver = &cicdmodels.EventReceiverModel{
			Active:                   m.EventReceiver.Active.ValueBool(),
			SCMType:                  m.EventReceiver.SCMType.ValueString(),
			WebhookTokenCredentialID: m.EventReceiver.WebhookTokenCredentialID.ValueString(),
		}
	}
	return req
}

// ---------------------------------------------------------------------------
// Data source model
// ---------------------------------------------------------------------------

// repositoryDSModel is the Terraform state model for the single-repository data source.
type repositoryDSModel struct {
	ID                         types.String          `tfsdk:"id"`
	Name                       types.String          `tfsdk:"name"`
	CloneURL                   types.String          `tfsdk:"clone_url"`
	CloneCredentialID          types.String          `tfsdk:"clone_credential_id"`
	CloudConnectorCredentialID types.String          `tfsdk:"cloud_connector_credential_id"`
	EventReceiver              *eventReceiverDSModel `tfsdk:"event_receiver"`
}

// eventReceiverDSModel is the data source variant of the event receiver nested object (all Computed).
type eventReceiverDSModel struct {
	Active                   types.Bool   `tfsdk:"active"`
	SCMType                  types.String `tfsdk:"scm_type"`
	WebhookID                types.String `tfsdk:"webhook_id"`
	WebhookTokenCredentialID types.String `tfsdk:"webhook_token_credential_id"`
}

func repositoryDSValueFrom(v cicdmodels.Repository) repositoryDSModel {
	m := repositoryDSModel{
		ID:                         types.StringValue(v.ID),
		Name:                       types.StringValue(v.Name),
		CloneURL:                   types.StringValue(v.CloneURL),
		CloneCredentialID:          optionalStringPtr(v.CloneCredentialID),
		CloudConnectorCredentialID: optionalStringPtr(v.CloudConnectorCredentialID),
	}
	if v.EventReceiver != nil {
		m.EventReceiver = &eventReceiverDSModel{
			Active:                   types.BoolValue(v.EventReceiver.Active),
			SCMType:                  types.StringValue(v.EventReceiver.SCMType),
			WebhookID:                types.StringValue(v.EventReceiver.WebhookID),
			WebhookTokenCredentialID: types.StringValue(v.EventReceiver.WebhookTokenCredentialID),
		}
	}
	return m
}

func (m repositoryResourceModel) toUpdateRequest(state repositoryResourceModel) cicdmodels.UpdateRepositoryRequest {
	req := cicdmodels.UpdateRepositoryRequest{
		ID:       state.ID.ValueString(),
		Name:     m.Name.ValueString(),
		CloneURL: m.CloneURL.ValueString(),
	}
	if !m.CloneCredentialID.IsNull() && !m.CloneCredentialID.IsUnknown() {
		v := m.CloneCredentialID.ValueString()
		req.CloneCredentialID = &v
	}
	if !m.CloudConnectorCredentialID.IsNull() && !m.CloudConnectorCredentialID.IsUnknown() {
		v := m.CloudConnectorCredentialID.ValueString()
		req.CloudConnectorCredentialID = &v
	}
	if m.EventReceiver != nil {
		webhookID := m.EventReceiver.WebhookID.ValueString()
		if webhookID == "" && state.EventReceiver != nil {
			webhookID = state.EventReceiver.WebhookID.ValueString()
		}
		req.EventReceiver = &cicdmodels.EventReceiverModel{
			Active:                   m.EventReceiver.Active.ValueBool(),
			SCMType:                  m.EventReceiver.SCMType.ValueString(),
			WebhookID:                webhookID,
			WebhookTokenCredentialID: m.EventReceiver.WebhookTokenCredentialID.ValueString(),
		}
	}
	return req
}

// ---------------------------------------------------------------------------
// DS list models
// ---------------------------------------------------------------------------

type repositoriesDSModel struct {
	ID     types.String `tfsdk:"id"`
	Values types.List   `tfsdk:"values"`
}

// repositoriesDSItemModel is one item in the values list.
type repositoriesDSItemModel struct {
	ID                         types.String          `tfsdk:"id"`
	Name                       types.String          `tfsdk:"name"`
	CloneURL                   types.String          `tfsdk:"clone_url"`
	CloneCredentialID          types.String          `tfsdk:"clone_credential_id"`
	CloudConnectorCredentialID types.String          `tfsdk:"cloud_connector_credential_id"`
	EventReceiver              *eventReceiverDSModel `tfsdk:"event_receiver"`
}

var repositoriesDSItemAttrTypes = map[string]attr.Type{
	"id":                            types.StringType,
	"name":                          types.StringType,
	"clone_url":                     types.StringType,
	"clone_credential_id":           types.StringType,
	"cloud_connector_credential_id": types.StringType,
	"event_receiver": types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"active":                      types.BoolType,
			"scm_type":                    types.StringType,
			"webhook_id":                  types.StringType,
			"webhook_token_credential_id": types.StringType,
		},
	},
}

var repositoriesDSItemType = types.ObjectType{AttrTypes: repositoriesDSItemAttrTypes}

func repositoriesDSItemsFrom(list []cicdmodels.Repository) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	items := make([]attr.Value, 0, len(list))
	for _, r := range list {
		item := repositoriesDSItemModel{
			ID:                         types.StringValue(r.ID),
			Name:                       types.StringValue(r.Name),
			CloneURL:                   types.StringValue(r.CloneURL),
			CloneCredentialID:          optionalStringPtr(r.CloneCredentialID),
			CloudConnectorCredentialID: optionalStringPtr(r.CloudConnectorCredentialID),
		}
		if r.EventReceiver != nil {
			item.EventReceiver = &eventReceiverDSModel{
				Active:                   types.BoolValue(r.EventReceiver.Active),
				SCMType:                  types.StringValue(r.EventReceiver.SCMType),
				WebhookID:                types.StringValue(r.EventReceiver.WebhookID),
				WebhookTokenCredentialID: types.StringValue(r.EventReceiver.WebhookTokenCredentialID),
			}
		}
		obj, d := types.ObjectValueFrom(context.Background(), repositoriesDSItemAttrTypes, item)
		diags.Append(d...)
		items = append(items, obj)
	}
	result, d := types.ListValue(repositoriesDSItemType, items)
	diags.Append(d...)
	return result, diags
}

// ---------------------------------------------------------------------------
// State models for repository jobs
// ---------------------------------------------------------------------------

type repositoryJobsDSModel struct {
	Repository types.String `tfsdk:"repository"`
	Values     types.List   `tfsdk:"values"`
}

type ansConfigDSModel struct {
	Active       types.Bool   `tfsdk:"active"`
	CredentialID types.String `tfsdk:"credential_id"`
	CustomTag    types.String `tfsdk:"custom_tag"`
}

type notificationConfigDSModel struct {
	ANS *ansConfigDSModel `tfsdk:"ans"`
}

type repositoryJobsDSItemModel struct {
	ID                        types.String               `tfsdk:"id"`
	Name                      types.String               `tfsdk:"name"`
	Active                    types.Bool                 `tfsdk:"active"`
	Description               types.String               `tfsdk:"description"`
	Pipeline                  types.String               `tfsdk:"pipeline"`
	PipelineVersion           types.String               `tfsdk:"pipeline_version"`
	PipelineParameters        types.Map                  `tfsdk:"pipeline_parameters"`
	BuildRetentionDays        types.Int64                `tfsdk:"build_retention_days"`
	MaxBuildsToKeep           types.Int64                `tfsdk:"max_builds_to_keep"`
	Branch                    types.String               `tfsdk:"branch"`
	RepositoryID              types.String               `tfsdk:"repository_id"`
	NotificationConfiguration *notificationConfigDSModel `tfsdk:"notification_configuration"`
}

var ansConfigAttrTypes = map[string]attr.Type{
	"active":        types.BoolType,
	"credential_id": types.StringType,
	"custom_tag":    types.StringType,
}

var notificationConfigAttrTypes = map[string]attr.Type{
	"ans": types.ObjectType{AttrTypes: ansConfigAttrTypes},
}

var repositoryJobsDSItemAttrTypes = map[string]attr.Type{
	"id":               types.StringType,
	"name":             types.StringType,
	"active":           types.BoolType,
	"description":      types.StringType,
	"pipeline":         types.StringType,
	"pipeline_version": types.StringType,
	"pipeline_parameters": types.MapType{
		ElemType: types.StringType,
	},
	"build_retention_days": types.Int64Type,
	"max_builds_to_keep":   types.Int64Type,
	"branch":               types.StringType,
	"repository_id":        types.StringType,
	"notification_configuration": types.ObjectType{
		AttrTypes: notificationConfigAttrTypes,
	},
}

var repositoryJobsDSItemType = types.ObjectType{AttrTypes: repositoryJobsDSItemAttrTypes}

func repositoryJobsDSItemsFrom(ctx context.Context, list []cicdmodels.Job) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	items := make([]attr.Value, 0, len(list))

	for _, j := range list {
		item, d := repositoryJobsDSItemFrom(ctx, j)
		diags.Append(d...)
		if diags.HasError() {
			return types.ListNull(repositoryJobsDSItemType), diags
		}
		items = append(items, item)
	}

	result, d := types.ListValue(repositoryJobsDSItemType, items)
	diags.Append(d...)
	return result, diags
}

func repositoryJobsDSItemFrom(ctx context.Context, j cicdmodels.Job) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	paramElems := make(map[string]attr.Value, len(j.PipelineParameters))
	for k, v := range j.PipelineParameters {
		paramElems[k] = types.StringValue(fmt.Sprintf("%v", v))
	}
	pipelineParams, d := types.MapValue(types.StringType, paramElems)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(repositoryJobsDSItemAttrTypes), diags
	}

	item := repositoryJobsDSItemModel{
		ID:                 types.StringValue(j.ID),
		Name:               types.StringValue(j.Name),
		Active:             types.BoolValue(j.Active),
		Description:        types.StringValue(j.Description),
		Pipeline:           types.StringValue(j.Pipeline),
		PipelineVersion:    types.StringValue(j.PipelineVersion),
		PipelineParameters: pipelineParams,
		BuildRetentionDays: types.Int64Value(int64(j.BuildRetentionDays)),
		MaxBuildsToKeep:    types.Int64Value(int64(j.MaxBuildsToKeep)),
		Branch:             types.StringValue(j.Branch),
		RepositoryID:       types.StringValue(j.RepositoryID),
	}

	if j.NotificationConfiguration != nil && j.NotificationConfiguration.ANS != nil {
		item.NotificationConfiguration = &notificationConfigDSModel{
			ANS: &ansConfigDSModel{
				Active:       types.BoolValue(j.NotificationConfiguration.ANS.Active),
				CredentialID: types.StringValue(j.NotificationConfiguration.ANS.CredentialID),
				CustomTag:    types.StringValue(j.NotificationConfiguration.ANS.CustomTag),
			},
		}
	} else {
		item.NotificationConfiguration = &notificationConfigDSModel{ANS: nil}
	}

	obj, d := types.ObjectValueFrom(ctx, repositoryJobsDSItemAttrTypes, item)
	diags.Append(d...)
	return obj, diags
}

// ---------------------------------------------------------------------------
// Event receiver data source model
// ---------------------------------------------------------------------------

type repositoryEventReceiverDSModel struct {
	Repository               types.String `tfsdk:"repository"`
	Active                   types.Bool   `tfsdk:"active"`
	SCMType                  types.String `tfsdk:"scm_type"`
	WebhookID                types.String `tfsdk:"webhook_id"`
	WebhookTokenCredentialID types.String `tfsdk:"webhook_token_credential_id"`
}

func repositoryEventReceiverDSValueFrom(repository string, v cicdmodels.EventReceiverModel) repositoryEventReceiverDSModel {
	webhookTokenCredID := types.StringNull()
	if v.WebhookTokenCredentialID != "" {
		webhookTokenCredID = types.StringValue(v.WebhookTokenCredentialID)
	}
	return repositoryEventReceiverDSModel{
		Repository:               types.StringValue(repository),
		Active:                   types.BoolValue(v.Active),
		SCMType:                  types.StringValue(v.SCMType),
		WebhookID:                types.StringValue(v.WebhookID),
		WebhookTokenCredentialID: webhookTokenCredID,
	}
}

// ---------------------------------------------------------------------------
// Webhook config data source model
// ---------------------------------------------------------------------------

type repositoryWebhookConfigDSModel struct {
	Repository types.String `tfsdk:"repository"`
	WebhookURI types.String `tfsdk:"webhook_uri"`
}

func repositoryWebhookConfigDSValueFrom(repository string, v cicdmodels.WebhookConfig) repositoryWebhookConfigDSModel {
	return repositoryWebhookConfigDSModel{
		Repository: types.StringValue(repository),
		WebhookURI: types.StringValue(v.WebhookURI),
	}
}
