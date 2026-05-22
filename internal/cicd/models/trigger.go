// internal/cicd/models/trigger.go

package cicdmodels

// TriggerTimer holds the schedule configuration for a timer-type trigger.
type TriggerTimer struct {
	Branch string `json:"branch,omitempty"`
	Cron   string `json:"cron,omitempty"`
}

// Trigger is the API response model for a single CI/CD job trigger.
type Trigger struct {
	ID         string        `json:"id"`
	APIVersion string        `json:"apiVersion,omitempty"`
	Type       string        `json:"type"`
	Timer      *TriggerTimer `json:"timer,omitempty"`
}

// CreateTriggerRequest is the body sent to POST /v2/jobs/{job}/triggers.
type CreateTriggerRequest struct {
	Type  string        `json:"type"`
	Timer *TriggerTimer `json:"timer,omitempty"`
}

// UpdateTriggerRequest is the body sent to PUT /v2/jobs/{job}/triggers/{id}.
type UpdateTriggerRequest struct {
	Type  string        `json:"type"`
	Timer *TriggerTimer `json:"timer,omitempty"`
}

// TriggerListResponse is the envelope returned by GET /v2/jobs/{job}/triggers.
type TriggerListResponse struct {
	Embedded *TriggerListEmbedded `json:"_embedded,omitempty"`
}

// TriggerListEmbedded holds the nested triggers array.
type TriggerListEmbedded struct {
	Triggers []Trigger `json:"triggers"`
}
