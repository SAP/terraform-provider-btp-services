// internal/cicd/client/facade_repositories.go

package cicdclient

import (
	"context"
	"fmt"

	cicdmodels "github.com/SAP/terraform-provider-sap-btp-services/internal/cicd/models"
)

type repositoriesFacade struct {
	hc *cicdHTTPClient
}

func newRepositoriesFacade(hc *cicdHTTPClient) repositoriesFacade {
	return repositoriesFacade{hc: hc}
}

// Create sends POST /v2/repositories.
// The API returns 201 with no body; a subsequent Get is required to read the assigned ID.
func (f *repositoriesFacade) Create(ctx context.Context, req cicdmodels.CreateRepositoryRequest) error {
	return f.hc.doPost(ctx, "/v2/repositories", req)
}

// Get sends GET /v2/repositories/{reference} where reference is the repository name or ID.
func (f *repositoriesFacade) Get(ctx context.Context, reference string) (*cicdmodels.Repository, error) {
	var result cicdmodels.Repository
	err := f.hc.doGet(ctx, fmt.Sprintf("/v2/repositories/%s", reference), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Update sends PUT /v2/repositories with the full model.
// The API returns 204 with no body.
func (f *repositoriesFacade) Update(ctx context.Context, req cicdmodels.UpdateRepositoryRequest) error {
	return f.hc.doPut(ctx, "/v2/repositories", req)
}

// Delete sends DELETE /v2/repositories/{reference}.
// The API returns 204 with no body.
func (f *repositoriesFacade) Delete(ctx context.Context, reference string) error {
	return f.hc.doDelete(ctx, fmt.Sprintf("/v2/repositories/%s", reference))
}

// List sends GET /v2/repositories and returns all repositories in the subaccount.
func (f *repositoriesFacade) List(ctx context.Context) ([]cicdmodels.Repository, error) {
	var result cicdmodels.RepositoryListResponse
	err := f.hc.doGet(ctx, "/v2/repositories", &result)
	if err != nil {
		return nil, err
	}
	if result.Embedded == nil {
		return []cicdmodels.Repository{}, nil
	}
	return result.Embedded.Repositories, nil
}

// GetEventReceiver sends GET /v2/repositories/{reference}/eventReceiver.
func (f *repositoriesFacade) GetEventReceiver(ctx context.Context, reference string) (*cicdmodels.EventReceiverModel, error) {
	var result cicdmodels.EventReceiverModel
	if err := f.hc.doGet(ctx, fmt.Sprintf("/v2/repositories/%s/eventReceiver", reference), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetWebhookConfig sends GET /v2/repositories/{reference}/webhookConfig.
func (f *repositoriesFacade) GetWebhookConfig(ctx context.Context, reference string) (*cicdmodels.WebhookConfig, error) {
	var result cicdmodels.WebhookConfig
	if err := f.hc.doGet(ctx, fmt.Sprintf("/v2/repositories/%s/webhookConfig", reference), &result); err != nil {
		return nil, err
	}
	return &result, nil
}
