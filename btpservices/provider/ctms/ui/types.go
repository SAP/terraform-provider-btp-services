package ctmsui

import "github.com/hashicorp/terraform-plugin-framework/types"

type uiURLModel struct {
	URL types.String `tfsdk:"url"`
}

type transportRequestLogsURLModel struct {
	TransportRequestID types.String `tfsdk:"transport_request_id"`
	NodeID             types.String `tfsdk:"node_id"`
	URL                types.String `tfsdk:"url"`
}
