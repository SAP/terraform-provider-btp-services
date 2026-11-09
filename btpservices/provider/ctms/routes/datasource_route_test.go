package ctmsroutes_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/testutils"
)

func TestDataSourceCTMSRoute(t *testing.T) {
	rec := testutils.SetupVCR(t, "datasource_route")
	config := `
resource "btpservice_ctms_node" "source" {
  name = "tf-ctms-ds-route-source"
}

resource "btpservice_ctms_node" "target" {
  name = "tf-ctms-ds-route-target"
}

resource "btpservice_ctms_route" "test" {
  name           = "tf-ctms-ds-route"
  source_node_id = btpservice_ctms_node.source.id
  target_node_id = btpservice_ctms_node.target.id
}

data "btpservice_ctms_route" "id" {
  id = btpservice_ctms_route.test.id
}

data "btpservice_ctms_route" "name" {
  name       = btpservice_ctms_route.test.name
  depends_on = [btpservice_ctms_route.test]
}
`

	resource.Test(
		t,
		rec.TestCase(
			resource.TestStep{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.btpservice_ctms_route.id", "id", "btpservice_ctms_route.test", "id"),
					resource.TestCheckResourceAttrPair("data.btpservice_ctms_route.name", "id", "btpservice_ctms_route.test", "id"),
				),
			},
			resource.TestStep{
				Config:   config,
				PlanOnly: true,
			},
		),
	)
}

func TestDataSourceCTMSRouteMissing(t *testing.T) {
	rec := testutils.SetupVCR(t, "datasource_route_missing")

	resource.Test(
		t,
		rec.TestCase(
			resource.TestStep{
				Config: `data "btpservice_ctms_route" "missing" {
  name = "tf-ctms-deliberately-missing-object"
}`,
				ExpectError: regexp.MustCompile("Error Reading CTMS"),
			},
		),
	)
}
