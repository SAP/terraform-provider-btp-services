package ctmsroutes_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/testutils"
)

func TestDataSourceCTMSRoutes(t *testing.T) {
	rec := testutils.SetupVCR(t, "datasource_routes")
	config := `
resource "btpservice_ctms_node" "source" {
  name = "tf-ctms-ds-routes-source"
}

resource "btpservice_ctms_node" "target" {
  name = "tf-ctms-ds-routes-target"
}

resource "btpservice_ctms_route" "test" {
  name           = "tf-ctms-ds-routes"
  source_node_id = btpservice_ctms_node.source.id
  target_node_id = btpservice_ctms_node.target.id
}

data "btpservice_ctms_routes" "all" {
  depends_on = [btpservice_ctms_route.test]
}
`

	resource.Test(
		t,
		rec.TestCase(
			resource.TestStep{
				Config: config,
				Check:  resource.TestCheckResourceAttrSet("data.btpservice_ctms_routes.all", "values.#"),
			},
			resource.TestStep{
				Config:   config,
				PlanOnly: true,
			},
		),
	)
}
