// btpservices/provider/cicd/jobs/types.go

package cicdjobs

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"gopkg.in/yaml.v3"

	cicdmodels "github.com/SAP/terraform-provider-sap-btp-services/internal/cicd/models"
)

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
