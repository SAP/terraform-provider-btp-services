package ctmsclient

import (
	"context"
	"fmt"
	"net/url"

	ctmsmodels "github.com/SAP/terraform-provider-btp-services/internal/ctms/models"
)

type destinationsFacade struct{ hc *ctmsHTTPClient }

func newDestinationsFacade(hc *ctmsHTTPClient) destinationsFacade {
	return destinationsFacade{hc: hc}
}

// List sends GET /v2/destinations and unwraps the destinations collection.
func (f *destinationsFacade) List(ctx context.Context) ([]ctmsmodels.DestinationBase, error) {
	var result ctmsmodels.DestinationsResponse
	if err := f.hc.doGet(ctx, "/v2/destinations", &result); err != nil {
		return nil, err
	}
	if result.Destinations == nil {
		return []ctmsmodels.DestinationBase{}, nil
	}
	return result.Destinations, nil
}

// Get sends GET /v2/destinations/{destinationName}.
// includeTenantID=false omits the optional query parameter, whose API default is false.
func (f *destinationsFacade) Get(ctx context.Context, name string, includeTenantID bool) (*ctmsmodels.DestinationExpanded, error) {
	path := "/v2/destinations/" + url.PathEscape(name)
	if includeTenantID {
		path += "?" + url.Values{"includeTenantId": {"true"}}.Encode()
	}
	var result ctmsmodels.DestinationExpanded
	if err := f.hc.doGet(ctx, path, &result); err != nil {
		return nil, err
	}
	if result.Name == nil {
		return nil, fmt.Errorf("GET %s response is missing destination name", path)
	}
	return &result, nil
}
