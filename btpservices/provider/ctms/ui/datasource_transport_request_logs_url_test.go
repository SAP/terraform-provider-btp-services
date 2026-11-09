package ctmsui_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/testutils"
)

func TestDataSourceCTMSTransportRequestLogsURL(t *testing.T) {
	rec := testutils.SetupVCR(t, "datasource_transport_request_logs_url", "test_transport_request_id", "test_node_id")
	config := fmt.Sprintf(
		`
data "btpservice_ctms_transport_request_logs_url" "request" {
  transport_request_id = %q
}

data "btpservice_ctms_transport_request_logs_url" "node" {
  transport_request_id = %q
  node_id              = %q
}
`,
		rec.Values["test_transport_request_id"],
		rec.Values["test_transport_request_id"],
		rec.Values["test_node_id"],
	)

	resource.Test(
		t,
		rec.TestCase(
			resource.TestStep{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.btpservice_ctms_transport_request_logs_url.request", "url"),
					resource.TestCheckResourceAttrSet("data.btpservice_ctms_transport_request_logs_url.node", "url"),
				),
			},
		),
	)
}
