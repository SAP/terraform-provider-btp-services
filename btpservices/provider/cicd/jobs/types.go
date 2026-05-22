// btpservices/provider/cicd/jobs/types.go

package cicdjobs

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	cicdmodels "github.com/SAP/terraform-provider-sap-btp-services/internal/cicd/models"
)

// buildTriggerIdentityModel is the identity of a build trigger resource.
type buildTriggerIdentityModel struct {
	Job types.String `tfsdk:"job"`
	ID  types.String `tfsdk:"id"`
}

// timerModel is the Terraform state model for the timer nested block.
type timerModel struct {
	Branch types.String `tfsdk:"branch"`
	Cron   types.String `tfsdk:"cron"`
}

// buildTriggerResourceModel is the Terraform state model for the build trigger resource.
type buildTriggerResourceModel struct {
	ID    types.String `tfsdk:"id"`
	Job   types.String `tfsdk:"job"`
	Type  types.String `tfsdk:"type"`
	Timer *timerModel  `tfsdk:"timer"`
}

func buildTriggerResourceValueFrom(job string, t cicdmodels.Trigger) buildTriggerResourceModel {
	m := buildTriggerResourceModel{
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

func (m buildTriggerResourceModel) toCreateRequest() cicdmodels.CreateTriggerRequest {
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

func (m buildTriggerResourceModel) toUpdateRequest() cicdmodels.UpdateTriggerRequest {
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
