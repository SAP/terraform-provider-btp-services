package ctmsclient

import (
	"context"
	"fmt"
	"net/url"

	ctmsmodels "github.com/SAP/terraform-provider-btp-services/internal/ctms/models"
)

type uiFacade struct{ hc *ctmsHTTPClient }

func newUIFacade(hc *ctmsHTTPClient) uiFacade {
	return uiFacade{hc: hc}
}

// GetURL sends GET /v2/tmsUiUrl and returns the tenant UI URL model.
func (f *uiFacade) GetURL(ctx context.Context) (*ctmsmodels.InlineResponse200, error) {
	var result ctmsmodels.InlineResponse200
	if err := f.hc.doGet(ctx, "/v2/tmsUiUrl", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTransportRequestLogsURL sends GET /v2/tmsUiUrl/transportRequestLogs.
// transportRequestID is required. A nil nodeID omits the optional query parameter.
// Both IDs are strings in this endpoint's schema.
func (f *uiFacade) GetTransportRequestLogsURL(
	ctx context.Context,
	transportRequestID string,
	nodeID *string,
) (*ctmsmodels.InlineResponse2001, error) {
	if transportRequestID == "" {
		return nil, fmt.Errorf("transportRequestID is required")
	}
	query := url.Values{"transportRequestId": {transportRequestID}}
	if nodeID != nil {
		query.Set("nodeId", *nodeID)
	}
	var result ctmsmodels.InlineResponse2001
	if err := f.hc.doGet(ctx, "/v2/tmsUiUrl/transportRequestLogs?"+query.Encode(), &result); err != nil {
		return nil, err
	}
	return &result, nil
}
