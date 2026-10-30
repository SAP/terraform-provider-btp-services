package ctmsclient

import (
	"context"
	"fmt"
	"net/http"

	ctmsmodels "github.com/SAP/terraform-provider-btp-services/internal/ctms/models"
)

type routesFacade struct{ hc *ctmsHTTPClient }

func newRoutesFacade(hc *ctmsHTTPClient) routesFacade {
	return routesFacade{hc: hc}
}

// Create sends POST /v2/routes and returns the created route, including its ID.
func (f *routesFacade) Create(ctx context.Context, req ctmsmodels.RouteCreateRequest) (*ctmsmodels.RouteResponse, error) {
	var result ctmsmodels.RouteResponse
	if err := f.hc.doWithBody(ctx, http.MethodPost, "/v2/routes", req, &result); err != nil {
		return nil, err
	}
	if result.Id == nil {
		return nil, fmt.Errorf("POST /v2/routes response is missing route ID")
	}
	return &result, nil
}

// Get sends GET /v2/routes/{routeId}. It needs no prior create or cached state.
func (f *routesFacade) Get(ctx context.Context, routeID int64) (*ctmsmodels.RouteResponse, error) {
	var result ctmsmodels.RouteResponse
	path := fmt.Sprintf("/v2/routes/%d", routeID)
	if err := f.hc.doGet(ctx, path, &result); err != nil {
		return nil, err
	}
	if result.Id == nil {
		return nil, fmt.Errorf("GET %s response is missing route ID", path)
	}
	if *result.Id != routeID {
		return nil, fmt.Errorf("GET %s returned a different route ID: %d", path, *result.Id)
	}
	return &result, nil
}

// Update sends PATCH /v2/routes/{routeId} as application/json and returns the updated route.
// Nil request fields are omitted; explicit values are encoded by pointer.
func (f *routesFacade) Update(ctx context.Context, routeID int64, req ctmsmodels.RoutePatchRequest) (*ctmsmodels.RouteResponse, error) {
	var result ctmsmodels.RouteResponse
	path := fmt.Sprintf("/v2/routes/%d", routeID)
	if err := f.hc.doWithBody(ctx, http.MethodPatch, path, req, &result); err != nil {
		return nil, err
	}
	if result.Id == nil {
		return nil, fmt.Errorf("PATCH %s response is missing route ID", path)
	}
	if *result.Id != routeID {
		return nil, fmt.Errorf("PATCH %s returned a different route ID: %d", path, *result.Id)
	}
	return &result, nil
}

// Delete sends DELETE /v2/routes/{routeId} and accepts an empty 204 response.
// A 404 is returned as NotFoundError; the resource decides how to handle it.
func (f *routesFacade) Delete(ctx context.Context, routeID int64) error {
	return f.hc.doDelete(ctx, fmt.Sprintf("/v2/routes/%d", routeID))
}

// List sends GET /v2/routes and unwraps the routes collection.
func (f *routesFacade) List(ctx context.Context) ([]ctmsmodels.RouteResponse, error) {
	var result ctmsmodels.RoutesResponse
	if err := f.hc.doGet(ctx, "/v2/routes", &result); err != nil {
		return nil, err
	}
	if result.Routes == nil {
		return []ctmsmodels.RouteResponse{}, nil
	}
	for _, item := range result.Routes {
		if item.Id == nil {
			return nil, fmt.Errorf("GET /v2/routes response contains a route without an ID")
		}
	}
	return result.Routes, nil
}
