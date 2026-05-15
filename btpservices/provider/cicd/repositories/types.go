// btpservices/provider/cicd/repositories/types.go

package cicdrepositories

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	cicdmodels "github.com/SAP/terraform-provider-sap-btp-services/internal/cicd/models"
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
