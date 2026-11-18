package ctmsui_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/testutils"
)

func TestDataSourceCTMSTransportRequestLogsURL(t *testing.T) {
	rec := testutils.SetupVCR(t, "datasource_transport_request_logs_url")
	config := testDataSourceCTMSTransportRequestLogsURLConfig()

	resource.Test(
		t,
		rec.TestCase(
			resource.TestStep{
				Config: config,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.btpservice_ctms_transport_request_logs_url.request", tfjsonpath.New("transport_request_id"), knownvalue.StringExact("201")),
					statecheck.ExpectKnownValue("data.btpservice_ctms_transport_request_logs_url.request", tfjsonpath.New("node_id"), knownvalue.Null()),
					statecheck.ExpectKnownValue("data.btpservice_ctms_transport_request_logs_url.request", tfjsonpath.New("url"), knownvalue.StringRegexp(regexp.MustCompile(`^https://.+`))),
				},
			},
			resource.TestStep{
				Config:   config,
				PlanOnly: true,
			},
		),
	)
}

func TestDataSourceCTMSTransportRequestLogsURLWithNode(t *testing.T) {
	rec := testutils.SetupVCR(t, "datasource_transport_request_logs_url_with_node")
	config := testDataSourceCTMSTransportRequestLogsURLWithNodeConfig()

	resource.Test(
		t,
		rec.TestCase(
			resource.TestStep{
				Config: config,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.btpservice_ctms_transport_request_logs_url.node", tfjsonpath.New("transport_request_id"), knownvalue.StringExact("201")),
					statecheck.ExpectKnownValue("data.btpservice_ctms_transport_request_logs_url.node", tfjsonpath.New("node_id"), knownvalue.StringExact("183")),
					statecheck.ExpectKnownValue("data.btpservice_ctms_transport_request_logs_url.node", tfjsonpath.New("url"), knownvalue.StringRegexp(regexp.MustCompile(`^https://.+`))),
				},
			},
			resource.TestStep{
				Config:   config,
				PlanOnly: true,
			},
		),
	)
}

func testDataSourceCTMSTransportRequestLogsURLConfig() string {
	return `
data "btpservice_ctms_transport_request_logs_url" "request" {
  transport_request_id = "201"
}
`
}

func testDataSourceCTMSTransportRequestLogsURLWithNodeConfig() string {
	return `
data "btpservice_ctms_transport_request_logs_url" "node" {
  transport_request_id = "201"
  node_id              = "183"
}
`
}
