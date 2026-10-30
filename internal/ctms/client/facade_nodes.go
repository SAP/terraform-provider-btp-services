package ctmsclient

import (
	"context"
	"fmt"
	"net/http"

	ctmsmodels "github.com/SAP/terraform-provider-btp-services/internal/ctms/models"
)

type nodesFacade struct{ hc *ctmsHTTPClient }

func newNodesFacade(hc *ctmsHTTPClient) nodesFacade {
	return nodesFacade{hc: hc}
}

// Create sends POST /v2/nodes and returns the created node, including its ID.
func (f *nodesFacade) Create(ctx context.Context, req ctmsmodels.NodeCreateRequest) (*ctmsmodels.NodeResponseTarget, error) {
	var result ctmsmodels.NodeResponseTarget
	if err := f.hc.doWithBody(ctx, http.MethodPost, "/v2/nodes", req, &result); err != nil {
		return nil, err
	}
	if result.Id == nil {
		return nil, fmt.Errorf("POST /v2/nodes response is missing node ID")
	}
	return &result, nil
}

// Get sends GET /v2/nodes/{nodeId}. It needs no prior create or cached state.
func (f *nodesFacade) Get(ctx context.Context, nodeID int64) (*ctmsmodels.NodeResponseTarget, error) {
	var result ctmsmodels.NodeResponseTarget
	path := fmt.Sprintf("/v2/nodes/%d", nodeID)
	if err := f.hc.doGet(ctx, path, &result); err != nil {
		return nil, err
	}
	if result.Id == nil {
		return nil, fmt.Errorf("GET %s response is missing node ID", path)
	}
	if *result.Id != nodeID {
		return nil, fmt.Errorf("GET %s returned a different node ID: %d", path, *result.Id)
	}
	return &result, nil
}

// Update sends PATCH /v2/nodes/{nodeId} as application/json and returns the updated node.
// Nil request fields are omitted; explicit values are encoded by pointer.
func (f *nodesFacade) Update(ctx context.Context, nodeID int64, req ctmsmodels.NodeUpdateRequest) (*ctmsmodels.NodeResponseTarget, error) {
	var result ctmsmodels.NodeResponseTarget
	path := fmt.Sprintf("/v2/nodes/%d", nodeID)
	if err := f.hc.doWithBody(ctx, http.MethodPatch, path, req, &result); err != nil {
		return nil, err
	}
	if result.Id == nil {
		return nil, fmt.Errorf("PATCH %s response is missing node ID", path)
	}
	if *result.Id != nodeID {
		return nil, fmt.Errorf("PATCH %s returned a different node ID: %d", path, *result.Id)
	}
	return &result, nil
}

// Delete sends DELETE /v2/nodes/{nodeId} and accepts an empty 204 response.
// A 404 is returned as NotFoundError; the resource decides how to handle it.
func (f *nodesFacade) Delete(ctx context.Context, nodeID int64) error {
	return f.hc.doDelete(ctx, fmt.Sprintf("/v2/nodes/%d", nodeID))
}

// List sends GET /v2/nodes and unwraps the nodes collection.
func (f *nodesFacade) List(ctx context.Context) ([]ctmsmodels.NodeResponseTarget, error) {
	var result ctmsmodels.NodesResponseTarget
	if err := f.hc.doGet(ctx, "/v2/nodes", &result); err != nil {
		return nil, err
	}
	if result.Nodes == nil {
		return []ctmsmodels.NodeResponseTarget{}, nil
	}
	for _, item := range result.Nodes {
		if item.Id == nil {
			return nil, fmt.Errorf("GET /v2/nodes response contains a node without an ID")
		}
	}
	return result.Nodes, nil
}
