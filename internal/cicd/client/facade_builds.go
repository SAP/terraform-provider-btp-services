// internal/cicd/client/facade_builds.go

package cicdclient

import (
	"context"
	"fmt"
	"net/url"

	cicdmodels "github.com/SAP/terraform-provider-btp-services/internal/cicd/models"
)

type buildsFacade struct {
	hc *cicdHTTPClient
}

func newBuildsFacade(hc *cicdHTTPClient) buildsFacade {
	return buildsFacade{hc: hc}
}

// Trigger sends POST /v2/jobs/{jobRef}/builds.
// The API returns 201 with no body on success.
// Returns a ConflictError (wrapping 409) if jobETag is stale.
func (f *buildsFacade) Trigger(ctx context.Context, jobRef string, req cicdmodels.BuildRequestDTO) error {
	return f.hc.doPost(ctx, fmt.Sprintf("/v2/jobs/%s/builds", url.PathEscape(jobRef)), req)
}

// Abort sends POST /v2/jobs/{jobRef}/builds/{buildRef}/abort.
// The API returns 202 with no body on success.
func (f *buildsFacade) Abort(ctx context.Context, jobRef, buildRef string) error {
	return f.hc.doPost(ctx,
		fmt.Sprintf("/v2/jobs/%s/builds/%s/abort", url.PathEscape(jobRef), url.PathEscape(buildRef)),
		struct{}{},
	)
}

// Delete sends DELETE /v2/jobs/{jobRef}/builds/{buildRef}.
// The API returns 204 with no body on success.
func (f *buildsFacade) Delete(ctx context.Context, jobRef, buildRef string) error {
	return f.hc.doDelete(ctx, fmt.Sprintf("/v2/jobs/%s/builds/%s", url.PathEscape(jobRef), url.PathEscape(buildRef)))
}

// ListBuilds sends GET /v2/jobs/{jobRef}/builds?filter={filter}.
// filter must be "latest" or "latestFinished".
func (f *buildsFacade) ListBuilds(ctx context.Context, jobRef, filter string) ([]cicdmodels.BuildStub, error) {
	path := fmt.Sprintf("/v2/jobs/%s/builds?filter=%s", url.PathEscape(jobRef), url.PathEscape(filter))
	var result cicdmodels.BuildListResponse
	if err := f.hc.doGet(ctx, path, &result); err != nil {
		return nil, err
	}
	if result.Embedded.Builds == nil {
		return []cicdmodels.BuildStub{}, nil
	}
	return result.Embedded.Builds, nil
}
