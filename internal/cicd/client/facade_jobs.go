// internal/cicd/client/facade_jobs.go

package cicdclient

import (
	"context"
	"fmt"
	"net/url"

	cicdmodels "github.com/SAP/terraform-provider-btp-services/internal/cicd/models"
)

type jobsFacade struct {
	hc *cicdHTTPClient
}

func newJobsFacade(hc *cicdHTTPClient) jobsFacade {
	return jobsFacade{hc: hc}
}

// Create sends POST /v2/jobs.
// The API returns 201 with no body; a subsequent Get is required to read the assigned ID.
func (f *jobsFacade) Create(ctx context.Context, req cicdmodels.CreateJobRequest) error {
	return f.hc.doPost(ctx, "/v2/jobs", req)
}

// Get sends GET /v2/jobs/{reference} where reference is the job name or ID.
func (f *jobsFacade) Get(ctx context.Context, reference string) (*cicdmodels.Job, error) {
	var result cicdmodels.Job
	err := f.hc.doGet(ctx, fmt.Sprintf("/v2/jobs/%s", url.PathEscape(reference)), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetWithETag sends GET /v2/jobs/{reference} and also returns the raw ETag response header
// value (e.g. W/"1"). The ETag is used to guard against triggering a build on a stale job.
func (f *jobsFacade) GetWithETag(ctx context.Context, reference string) (*cicdmodels.Job, string, error) {
	var result cicdmodels.Job
	etag, err := f.hc.doGetWithETag(ctx, fmt.Sprintf("/v2/jobs/%s", url.PathEscape(reference)), &result)
	if err != nil {
		return nil, "", err
	}
	result.ETag = etag
	return &result, etag, nil
}

// Update sends PUT /v2/jobs (full replace, job identified by name in request body).
// The API returns 204 with no body.
func (f *jobsFacade) Update(ctx context.Context, req cicdmodels.UpdateJobRequest) error {
	return f.hc.doPut(ctx, "/v2/jobs", req)
}

// Delete sends DELETE /v2/jobs/{reference}.
// The API returns 204 with no body.
func (f *jobsFacade) Delete(ctx context.Context, reference string) error {
	return f.hc.doDelete(ctx, fmt.Sprintf("/v2/jobs/%s", url.PathEscape(reference)))
}

// List sends GET /v2/jobs and returns all jobs across all repositories.
func (f *jobsFacade) List(ctx context.Context) ([]cicdmodels.Job, error) {
	var result cicdmodels.JobListResponse
	if err := f.hc.doGet(ctx, "/v2/jobs", &result); err != nil {
		return nil, err
	}
	if result.Embedded == nil {
		return []cicdmodels.Job{}, nil
	}
	return result.Embedded.Jobs, nil
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
// POST /v2/jobs/{job}/triggers returns 201 with no body and no Location header (per API spec),
// so we list before and after to identify the newly created trigger by ID diff.
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
