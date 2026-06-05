// internal/cicd/models/build.go

package cicdmodels

// BuildRequestDTO is the request body for POST /v2/jobs/{ref}/builds.
type BuildRequestDTO struct {
	CommitToBeBuilt string           `json:"commitToBeBuilt,omitempty"`
	JobETag         string           `json:"jobETag,omitempty"`
	Parameters      []BuildParameter `json:"parameters,omitempty"`
}

// BuildParameter is a per-build runtime parameter passed to the pipeline.
type BuildParameter struct {
	Name       string `json:"name"`
	Value      string `json:"value"`
	Visibility string `json:"visibility,omitempty"` // PUBLIC | RESTRICTED
}

// BuildStub is a single item returned by GET /v2/jobs/{ref}/builds?filter=...
type BuildStub struct {
	ID     string `json:"id"`
	Number int    `json:"number"`
}

// BuildListResponse is the HAL envelope for the builds list endpoint.
type BuildListResponse struct {
	Embedded struct {
		Builds []BuildStub `json:"builds"`
	} `json:"_embedded"`
}
