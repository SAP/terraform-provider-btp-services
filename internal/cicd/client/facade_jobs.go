// internal/cicd/client/facade_jobs.go

package cicdclient

import (
	"context"
	"fmt"
	"net/url"

	cicdmodels "github.com/SAP/terraform-provider-sap-btp-services/internal/cicd/models"
)

type jobsFacade struct {
	hc *cicdHTTPClient
}

func newJobsFacade(hc *cicdHTTPClient) jobsFacade {
	return jobsFacade{hc: hc}
}

// ListByRepository sends GET /v2/repositories/{reference}/jobs and returns all jobs for the repository.
// reference is the repository name or ID.
func (f *jobsFacade) ListByRepository(ctx context.Context, repositoryReference string) ([]cicdmodels.Job, error) {
	var result cicdmodels.JobListResponse
	err := f.hc.doGet(ctx, fmt.Sprintf("/v2/repositories/%s/jobs", url.PathEscape(repositoryReference)), &result)
	if err != nil {
		return nil, err
	}
	if result.Embedded == nil {
		return []cicdmodels.Job{}, nil
	}
	return result.Embedded.Jobs, nil
}

// CreateTrigger creates a new trigger under a job and returns the created trigger.
// POST /v2/jobs/{job}/triggers returns 201 with no body, so we list before and after to find the new ID.
func (f *jobsFacade) CreateTrigger(ctx context.Context, jobRef string, req cicdmodels.CreateTriggerRequest) (*cicdmodels.Trigger, error) {
	before, err := f.ListTriggers(ctx, jobRef)
	if err != nil {
		return nil, fmt.Errorf("list triggers before create: %w", err)
	}
	existingIDs := make(map[string]struct{}, len(before))
	for _, t := range before {
		existingIDs[t.ID] = struct{}{}
	}

	if err := f.hc.doPost(ctx, fmt.Sprintf("/v2/jobs/%s/triggers", url.PathEscape(jobRef)), req); err != nil {
		return nil, err
	}

	after, err := f.ListTriggers(ctx, jobRef)
	if err != nil {
		return nil, fmt.Errorf("list triggers after create: %w", err)
	}
	for i := range after {
		if _, exists := existingIDs[after[i].ID]; !exists {
			return &after[i], nil
		}
	}
	return nil, fmt.Errorf("created trigger not found in job %s", jobRef)
}

// GetTrigger sends GET /v2/jobs/{job}/triggers/{id}.
func (f *jobsFacade) GetTrigger(ctx context.Context, jobRef, triggerID string) (*cicdmodels.Trigger, error) {
	var result cicdmodels.Trigger
	if err := f.hc.doGet(ctx, fmt.Sprintf("/v2/jobs/%s/triggers/%s", url.PathEscape(jobRef), url.PathEscape(triggerID)), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateTrigger sends PUT /v2/jobs/{job}/triggers/{id}.
// The API returns 204 with no body.
func (f *jobsFacade) UpdateTrigger(ctx context.Context, jobRef, triggerID string, req cicdmodels.UpdateTriggerRequest) error {
	return f.hc.doPut(ctx, fmt.Sprintf("/v2/jobs/%s/triggers/%s", url.PathEscape(jobRef), url.PathEscape(triggerID)), req)
}

// DeleteTrigger sends DELETE /v2/jobs/{job}/triggers/{id}.
// The API returns 204 with no body.
func (f *jobsFacade) DeleteTrigger(ctx context.Context, jobRef, triggerID string) error {
	return f.hc.doDelete(ctx, fmt.Sprintf("/v2/jobs/%s/triggers/%s", url.PathEscape(jobRef), url.PathEscape(triggerID)))
}

// ListTriggers sends GET /v2/jobs/{job}/triggers and returns all triggers for the job.
func (f *jobsFacade) ListTriggers(ctx context.Context, jobRef string) ([]cicdmodels.Trigger, error) {
	var result cicdmodels.TriggerListResponse
	if err := f.hc.doGet(ctx, fmt.Sprintf("/v2/jobs/%s/triggers", url.PathEscape(jobRef)), &result); err != nil {
		return nil, err
	}
	if result.Embedded == nil {
		return []cicdmodels.Trigger{}, nil
	}
	return result.Embedded.Triggers, nil
}
