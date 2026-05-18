// internal/cicd/client/facade_credentials.go

package cicdclient

import (
	"context"
	"fmt"
	"net/url"

	cicdmodels "github.com/SAP/terraform-provider-sap-btp-services/internal/cicd/models"
)

type credentialsFacade struct {
	hc *cicdHTTPClient
}

func newCredentialsFacade(hc *cicdHTTPClient) credentialsFacade {
	return credentialsFacade{hc: hc}
}

// Create sends POST /v2/credentials.
// The API returns 201 with no body; a subsequent Get is required to read the assigned ID.
func (f *credentialsFacade) Create(ctx context.Context, req cicdmodels.CreateCredentialRequest) error {
	return f.hc.doPost(ctx, "/v2/credentials", req)
}

// Get sends GET /v2/credentials/{reference} where reference is the credential name or ID.
func (f *credentialsFacade) Get(ctx context.Context, reference string) (*cicdmodels.Credential, error) {
	var result cicdmodels.Credential
	err := f.hc.doGet(ctx, fmt.Sprintf("/v2/credentials/%s", reference), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Update sends PUT /v2/credentials/{reference}.
// The API returns 204 with no body.
func (f *credentialsFacade) Update(ctx context.Context, reference string, req cicdmodels.UpdateCredentialRequest) error {
	return f.hc.doPut(ctx, fmt.Sprintf("/v2/credentials/%s", reference), req)
}

// Delete sends DELETE /v2/credentials/{reference}.
// The API returns 204 with no body.
func (f *credentialsFacade) Delete(ctx context.Context, reference string) error {
	return f.hc.doDelete(ctx, fmt.Sprintf("/v2/credentials/%s", reference))
}

// Patch sends PATCH /v2/credentials/{reference} with merge-patch semantics.
// Only fields set in req are updated; omitted fields are left unchanged by the API.
func (f *credentialsFacade) Patch(ctx context.Context, reference string, req cicdmodels.PatchCredentialRequest) error {
	return f.hc.doPatch(ctx, fmt.Sprintf("/v2/credentials/%s", reference), req)
}

// List sends GET /v2/credentials and returns all credentials in the subaccount.
func (f *credentialsFacade) List(ctx context.Context) ([]cicdmodels.Credential, error) {
	var result cicdmodels.CredentialListResponse
	err := f.hc.doGet(ctx, "/v2/credentials", &result)
	if err != nil {
		return nil, err
	}
	if result.Embedded == nil {
		return []cicdmodels.Credential{}, nil
	}
	return result.Embedded.Credentials, nil
}

// GetUsages sends GET /v2/credentials/{reference}/usages and returns the list of
// jobs and repositories that reference the given credential.
// usertype is optional — pass "" to return all usage types, or "job"/"repository" to filter.
func (f *credentialsFacade) GetUsages(ctx context.Context, reference, usertype string) ([]cicdmodels.CredentialUsage, error) {
	path := fmt.Sprintf("/v2/credentials/%s/usages", reference)
	if usertype != "" {
		q := url.Values{}
		q.Set("usertype", usertype)
		path = path + "?" + q.Encode()
	}
	var result cicdmodels.CredentialUsageListResponse
	if err := f.hc.doGet(ctx, path, &result); err != nil {
		return nil, err
	}
	if result.Embedded == nil {
		return []cicdmodels.CredentialUsage{}, nil
	}
	return result.Embedded.Usages, nil
}

// GetJobCredentials sends GET /v2/jobs/{jobReference}/credentials and returns the list
// of credential IDs that a build of the job is allowed to use.
func (f *credentialsFacade) GetJobCredentials(ctx context.Context, jobReference string) ([]string, error) {
	var result cicdmodels.JobCredentialListResponse
	if err := f.hc.doGet(ctx, fmt.Sprintf("/v2/jobs/%s/credentials", jobReference), &result); err != nil {
		return nil, err
	}
	if result.IDs == nil {
		return []string{}, nil
	}
	return result.IDs, nil
}
