// btpservices/provider/cicd/jobs/types.go

package cicdjobs

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gopkg.in/yaml.v3"

	cicdmodels "github.com/SAP/terraform-provider-btp-services/internal/cicd/models"
)

// jobIdentityModel is the identity of a job resource.
type jobIdentityModel struct {
	ID types.String `tfsdk:"id"`
}

// jobResourceModel is the Terraform state model for the job resource.
type jobResourceModel struct {
	ID                        types.String                    `tfsdk:"id"`
	Name                      types.String                    `tfsdk:"name"`
	Description               types.String                    `tfsdk:"description"`
	Active                    types.Bool                      `tfsdk:"active"`
	Pipeline                  types.String                    `tfsdk:"pipeline"`
	PipelineVersion           types.String                    `tfsdk:"pipeline_version"`
	PipelineParameters        types.String                    `tfsdk:"pipeline_parameters"`
	BuildRetentionDays        types.Int64                     `tfsdk:"build_retention_days"`
	MaxBuildsToKeep           types.Int64                     `tfsdk:"max_builds_to_keep"`
	Branch                    types.String                    `tfsdk:"branch"`
	RepositoryID              types.String                    `tfsdk:"repository_id"`
	NotificationConfiguration *notificationConfigurationModel `tfsdk:"notification_configuration"`
}

type notificationConfigurationModel struct {
	ANS *ansConfigurationModel `tfsdk:"ans"`
}

type ansConfigurationModel struct {
	Active       types.Bool   `tfsdk:"active"`
	CredentialID types.String `tfsdk:"credential_id"`
	CustomTag    types.String `tfsdk:"custom_tag"`
}

// yamlToMap parses a YAML string into map[string]any for the API request.
func yamlToMap(yamlStr string) (map[string]any, error) {
	var m map[string]any
	if err := yaml.Unmarshal([]byte(yamlStr), &m); err != nil {
		return nil, fmt.Errorf("invalid YAML in pipeline_parameters: %w", err)
	}
	if m == nil {
		m = map[string]any{}
	}
	return m, nil
}

// mapToYAML serializes a map to a canonical YAML string (sorted keys).
// Used only for the import path where no prior user YAML exists.
func mapToYAML(m map[string]any) (string, error) {
	b, err := yaml.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("failed to serialize pipelineParameters to YAML: %w", err)
	}
	return string(b), nil
}

// jobResourceValueFrom maps an API Job response to the Terraform state model.
//
// priorYAML is the pipeline_parameters value already in state (what the user
// wrote). We always preserve it on normal CRUD paths because the API
// re-serialises pipelineParameters with sorted keys and different indentation —
// storing the API form would always differ from what the plan expected and
// cause "provider produced inconsistent result" errors.
//
// On import isImport is true. In that case we fall back to serialising
// the API response to canonical YAML so the state is populated correctly.
func jobResourceValueFrom(v cicdmodels.Job, priorYAML string, isImport bool) (jobResourceModel, error) {
	pipelineParams := priorYAML
	if isImport {
		// Import path: no prior user YAML — serialize the API response.
		canonical, err := mapToYAML(v.PipelineParameters)
		if err != nil {
			return jobResourceModel{}, err
		}
		pipelineParams = canonical
	}

	var notifConfig *notificationConfigurationModel
	if v.NotificationConfiguration != nil {
		notifConfig = &notificationConfigurationModel{}
		if v.NotificationConfiguration.ANS != nil {
			notifConfig.ANS = &ansConfigurationModel{
				Active:       types.BoolValue(v.NotificationConfiguration.ANS.Active),
				CredentialID: types.StringValue(v.NotificationConfiguration.ANS.CredentialID),
				CustomTag:    types.StringValue(v.NotificationConfiguration.ANS.CustomTag),
			}
		}
	}

	return jobResourceModel{
		ID:                        types.StringValue(v.ID),
		Name:                      types.StringValue(v.Name),
		Description:               types.StringValue(v.Description),
		Active:                    types.BoolValue(v.Active),
		Pipeline:                  types.StringValue(v.Pipeline),
		PipelineVersion:           types.StringValue(v.PipelineVersion),
		PipelineParameters:        types.StringValue(pipelineParams),
		BuildRetentionDays:        types.Int64Value(int64(v.BuildRetentionDays)),
		MaxBuildsToKeep:           types.Int64Value(int64(v.MaxBuildsToKeep)),
		Branch:                    types.StringValue(v.Branch),
		RepositoryID:              types.StringValue(v.RepositoryID),
		NotificationConfiguration: notifConfig,
	}, nil
}

func (m jobResourceModel) toCreateRequest() (cicdmodels.CreateJobRequest, error) {
	params, err := yamlToMap(m.PipelineParameters.ValueString())
	if err != nil {
		return cicdmodels.CreateJobRequest{}, err
	}
	return cicdmodels.CreateJobRequest{
		Name:                      m.Name.ValueString(),
		Description:               m.Description.ValueString(),
		Active:                    m.Active.ValueBool(),
		Pipeline:                  m.Pipeline.ValueString(),
		PipelineVersion:           m.PipelineVersion.ValueString(),
		PipelineParameters:        params,
		BuildRetentionDays:        m.BuildRetentionDays.ValueInt64(),
		MaxBuildsToKeep:           m.MaxBuildsToKeep.ValueInt64(),
		Branch:                    m.Branch.ValueString(),
		RepositoryID:              m.RepositoryID.ValueString(),
		NotificationConfiguration: m.toAPINotificationConfiguration(),
	}, nil
}

func (m jobResourceModel) toUpdateRequest(id string) (cicdmodels.UpdateJobRequest, error) {
	params, err := yamlToMap(m.PipelineParameters.ValueString())
	if err != nil {
		return cicdmodels.UpdateJobRequest{}, err
	}
	return cicdmodels.UpdateJobRequest{
		ID:                        id,
		Name:                      m.Name.ValueString(),
		Description:               m.Description.ValueString(),
		Active:                    m.Active.ValueBool(),
		Pipeline:                  m.Pipeline.ValueString(),
		PipelineVersion:           m.PipelineVersion.ValueString(),
		PipelineParameters:        params,
		BuildRetentionDays:        m.BuildRetentionDays.ValueInt64(),
		MaxBuildsToKeep:           m.MaxBuildsToKeep.ValueInt64(),
		Branch:                    m.Branch.ValueString(),
		RepositoryID:              m.RepositoryID.ValueString(),
		NotificationConfiguration: m.toAPINotificationConfiguration(),
	}, nil
}

func (m jobResourceModel) toAPINotificationConfiguration() *cicdmodels.NotificationConfiguration {
	if m.NotificationConfiguration == nil {
		return nil
	}
	notif := &cicdmodels.NotificationConfiguration{}
	if m.NotificationConfiguration.ANS != nil {
		notif.ANS = &cicdmodels.AnsConfiguration{
			Active:       m.NotificationConfiguration.ANS.Active.ValueBool(),
			CredentialID: m.NotificationConfiguration.ANS.CredentialID.ValueString(),
			CustomTag:    m.NotificationConfiguration.ANS.CustomTag.ValueString(),
		}
	}
	return notif
}

// ---------------------------------------------------------------------------
// Data source model
// ---------------------------------------------------------------------------

// jobDSModel is the Terraform state model for the single-job data source.
type jobDSModel struct {
	ID                        types.String                    `tfsdk:"id"`
	Name                      types.String                    `tfsdk:"name"`
	Description               types.String                    `tfsdk:"description"`
	Active                    types.Bool                      `tfsdk:"active"`
	Pipeline                  types.String                    `tfsdk:"pipeline"`
	PipelineVersion           types.String                    `tfsdk:"pipeline_version"`
	PipelineParameters        types.String                    `tfsdk:"pipeline_parameters"`
	BuildRetentionDays        types.Int64                     `tfsdk:"build_retention_days"`
	MaxBuildsToKeep           types.Int64                     `tfsdk:"max_builds_to_keep"`
	Branch                    types.String                    `tfsdk:"branch"`
	RepositoryID              types.String                    `tfsdk:"repository_id"`
	ETag                      types.String                    `tfsdk:"etag"`
	NotificationConfiguration *notificationConfigurationModel `tfsdk:"notification_configuration"`
}

// jobDSValueFrom maps an API Job response to the data source state model.
// pipeline_parameters is serialised to canonical YAML.
// etag is the raw ETag response header value (e.g. W/"1").
func jobDSValueFrom(v cicdmodels.Job, etag string) (jobDSModel, error) {
	yamlStr, err := mapToYAML(v.PipelineParameters)
	if err != nil {
		return jobDSModel{}, err
	}

	var notifConfig *notificationConfigurationModel
	if v.NotificationConfiguration != nil {
		notifConfig = &notificationConfigurationModel{}
		if v.NotificationConfiguration.ANS != nil {
			notifConfig.ANS = &ansConfigurationModel{
				Active:       types.BoolValue(v.NotificationConfiguration.ANS.Active),
				CredentialID: types.StringValue(v.NotificationConfiguration.ANS.CredentialID),
				CustomTag:    types.StringValue(v.NotificationConfiguration.ANS.CustomTag),
			}
		}
	}

	return jobDSModel{
		ID:                        types.StringValue(v.ID),
		Name:                      types.StringValue(v.Name),
		Description:               types.StringValue(v.Description),
		Active:                    types.BoolValue(v.Active),
		Pipeline:                  types.StringValue(v.Pipeline),
		PipelineVersion:           types.StringValue(v.PipelineVersion),
		PipelineParameters:        types.StringValue(yamlStr),
		BuildRetentionDays:        types.Int64Value(int64(v.BuildRetentionDays)),
		MaxBuildsToKeep:           types.Int64Value(int64(v.MaxBuildsToKeep)),
		Branch:                    types.StringValue(v.Branch),
		RepositoryID:              types.StringValue(v.RepositoryID),
		ETag:                      types.StringValue(etag),
		NotificationConfiguration: notifConfig,
	}, nil
}

// triggerIdentityModel is the identity of a trigger resource.
type triggerIdentityModel struct {
	Job types.String `tfsdk:"job"`
	ID  types.String `tfsdk:"id"`
}

// timerModel is the Terraform state model for the timer nested block.
type timerModel struct {
	Branch types.String `tfsdk:"branch"`
	Cron   types.String `tfsdk:"cron"`
}

// triggerResourceModel is the Terraform state model for the trigger resource.
type triggerResourceModel struct {
	ID    types.String `tfsdk:"id"`
	Job   types.String `tfsdk:"job"`
	Type  types.String `tfsdk:"type"`
	Timer *timerModel  `tfsdk:"timer"`
}

func triggerResourceValueFrom(job string, t cicdmodels.Trigger) triggerResourceModel {
	m := triggerResourceModel{
		ID:   types.StringValue(t.ID),
		Job:  types.StringValue(job),
		Type: types.StringValue(t.Type),
	}
	if t.Timer != nil {
		m.Timer = &timerModel{
			Branch: types.StringValue(t.Timer.Branch),
			Cron:   types.StringValue(t.Timer.Cron),
		}
	}
	return m
}

func (m triggerResourceModel) toCreateRequest() cicdmodels.CreateTriggerRequest {
	req := cicdmodels.CreateTriggerRequest{
		Type: m.Type.ValueString(),
	}
	if m.Timer != nil {
		req.Timer = &cicdmodels.TriggerTimer{
			Branch: m.Timer.Branch.ValueString(),
			Cron:   m.Timer.Cron.ValueString(),
		}
	}
	return req
}

// jobsDSModel is the Terraform state model for the jobs (plural) data source.
type jobsDSModel struct {
	ID     types.String `tfsdk:"id"`
	Values types.List   `tfsdk:"values"`
}

// jobsDSItemModel is one item in the values list.
type jobsDSItemModel struct {
	ID                        types.String                    `tfsdk:"id"`
	Name                      types.String                    `tfsdk:"name"`
	Description               types.String                    `tfsdk:"description"`
	Active                    types.Bool                      `tfsdk:"active"`
	Pipeline                  types.String                    `tfsdk:"pipeline"`
	PipelineVersion           types.String                    `tfsdk:"pipeline_version"`
	PipelineParameters        types.String                    `tfsdk:"pipeline_parameters"`
	BuildRetentionDays        types.Int64                     `tfsdk:"build_retention_days"`
	MaxBuildsToKeep           types.Int64                     `tfsdk:"max_builds_to_keep"`
	Branch                    types.String                    `tfsdk:"branch"`
	RepositoryID              types.String                    `tfsdk:"repository_id"`
	NotificationConfiguration *notificationConfigurationModel `tfsdk:"notification_configuration"`
}

var jobsDSANSAttrTypes = map[string]attr.Type{
	"active":        types.BoolType,
	"credential_id": types.StringType,
	"custom_tag":    types.StringType,
}

var jobsDSNotificationAttrTypes = map[string]attr.Type{
	"ans": types.ObjectType{AttrTypes: jobsDSANSAttrTypes},
}

var jobsDSItemAttrTypes = map[string]attr.Type{
	"id":                   types.StringType,
	"name":                 types.StringType,
	"description":          types.StringType,
	"active":               types.BoolType,
	"pipeline":             types.StringType,
	"pipeline_version":     types.StringType,
	"pipeline_parameters":  types.StringType,
	"build_retention_days": types.Int64Type,
	"max_builds_to_keep":   types.Int64Type,
	"branch":               types.StringType,
	"repository_id":        types.StringType,
	"notification_configuration": types.ObjectType{
		AttrTypes: jobsDSNotificationAttrTypes,
	},
}

var jobsDSItemType = types.ObjectType{AttrTypes: jobsDSItemAttrTypes}

func jobsDSItemsFrom(ctx context.Context, list []cicdmodels.Job) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	items := make([]attr.Value, 0, len(list))

	for _, j := range list {
		yamlStr, err := mapToYAML(j.PipelineParameters)
		if err != nil {
			diags.AddError("Error Serializing Pipeline Parameters", err.Error())
			return types.ListNull(jobsDSItemType), diags
		}

		item := jobsDSItemModel{
			ID:                 types.StringValue(j.ID),
			Name:               types.StringValue(j.Name),
			Description:        types.StringValue(j.Description),
			Active:             types.BoolValue(j.Active),
			Pipeline:           types.StringValue(j.Pipeline),
			PipelineVersion:    types.StringValue(j.PipelineVersion),
			PipelineParameters: types.StringValue(yamlStr),
			BuildRetentionDays: types.Int64Value(int64(j.BuildRetentionDays)),
			MaxBuildsToKeep:    types.Int64Value(int64(j.MaxBuildsToKeep)),
			Branch:             types.StringValue(j.Branch),
			RepositoryID:       types.StringValue(j.RepositoryID),
		}

		if j.NotificationConfiguration != nil && j.NotificationConfiguration.ANS != nil {
			item.NotificationConfiguration = &notificationConfigurationModel{
				ANS: &ansConfigurationModel{
					Active:       types.BoolValue(j.NotificationConfiguration.ANS.Active),
					CredentialID: types.StringValue(j.NotificationConfiguration.ANS.CredentialID),
					CustomTag:    types.StringValue(j.NotificationConfiguration.ANS.CustomTag),
				},
			}
		} else {
			item.NotificationConfiguration = &notificationConfigurationModel{ANS: nil}
		}

		obj, d := types.ObjectValueFrom(ctx, jobsDSItemAttrTypes, item)
		diags.Append(d...)
		if diags.HasError() {
			return types.ListNull(jobsDSItemType), diags
		}
		items = append(items, obj)
	}

	result, d := types.ListValue(jobsDSItemType, items)
	diags.Append(d...)
	return result, diags
}

func (m triggerResourceModel) toUpdateRequest() cicdmodels.UpdateTriggerRequest {
	req := cicdmodels.UpdateTriggerRequest{
		Type: m.Type.ValueString(),
	}
	if m.Timer != nil {
		req.Timer = &cicdmodels.TriggerTimer{
			Branch: m.Timer.Branch.ValueString(),
			Cron:   m.Timer.Cron.ValueString(),
		}
	}
	return req
}

// triggerDSItem is a single trigger inside the btpservice_cicd_triggers data source values list.
type triggerDSItem struct {
	ID    types.String `tfsdk:"id"`
	Type  types.String `tfsdk:"type"`
	Timer *timerModel  `tfsdk:"timer"`
}

// triggersDSModel is the Terraform state model for the btpservice_cicd_triggers data source.
type triggersDSModel struct {
	ID     types.String    `tfsdk:"id"`
	Job    types.String    `tfsdk:"job"`
	Values []triggerDSItem `tfsdk:"values"`
}

// triggerDSModel is the Terraform state model for the btpservice_cicd_trigger data source.
type triggerDSModel struct {
	ID    types.String `tfsdk:"id"`
	Job   types.String `tfsdk:"job"`
	Type  types.String `tfsdk:"type"`
	Timer *timerModel  `tfsdk:"timer"`
}

func triggerDSItemFrom(t cicdmodels.Trigger) triggerDSItem {
	item := triggerDSItem{
		ID:   types.StringValue(t.ID),
		Type: types.StringValue(t.Type),
	}
	if t.Timer != nil {
		item.Timer = &timerModel{
			Branch: types.StringValue(t.Timer.Branch),
			Cron:   types.StringValue(t.Timer.Cron),
		}
	}
	return item
}

func triggerDSValueFrom(job string, t cicdmodels.Trigger) triggerDSModel {
	m := triggerDSModel{
		ID:   types.StringValue(t.ID),
		Job:  types.StringValue(job),
		Type: types.StringValue(t.Type),
	}
	if t.Timer != nil {
		m.Timer = &timerModel{
			Branch: types.StringValue(t.Timer.Branch),
			Cron:   types.StringValue(t.Timer.Cron),
		}
	}
	return m
}

// ---------------------------------------------------------------------------
// Builds data source models
// ---------------------------------------------------------------------------

// buildsDSModel is the Terraform state model for btpservice_cicd_builds.
type buildsDSModel struct {
	ID     types.String  `tfsdk:"id"`
	Job    types.String  `tfsdk:"job"`
	Filter types.String  `tfsdk:"filter"`
	Builds []buildDSItem `tfsdk:"builds"`
}

// buildDSItem is a single build inside the builds list.
type buildDSItem struct {
	ID     types.String `tfsdk:"id"`
	Number types.Int64  `tfsdk:"number"`
}

func buildsDSItemsFrom(builds []cicdmodels.BuildStub) []buildDSItem {
	items := make([]buildDSItem, 0, len(builds))
	for _, b := range builds {
		items = append(items, buildDSItem{
			ID:     types.StringValue(b.ID),
			Number: types.Int64Value(int64(b.Number)),
		})
	}
	return items
}
