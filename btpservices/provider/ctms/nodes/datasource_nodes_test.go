package ctmsnodes_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/testutils"
)

func TestDataSourceCTMSNodes(t *testing.T) {
	rec := testutils.SetupVCR(t, "datasource_nodes")
	config := `
resource "btpservice_ctms_node" "test" {
  name = "tf-ctms-ds-nodes"
}

data "btpservice_ctms_nodes" "all" {
  depends_on = [btpservice_ctms_node.test]
}
`

	resource.Test(
		t,
		rec.TestCase(
			resource.TestStep{
				Config: config,
				Check:  resource.TestCheckResourceAttrSet("data.btpservice_ctms_nodes.all", "values.#"),
			},
			resource.TestStep{
				Config:   config,
				PlanOnly: true,
			},
		),
	)
}
