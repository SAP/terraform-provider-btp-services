package ctmsnodes_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/testutils"
)

func TestDataSourceCTMSNode(t *testing.T) {
	rec := testutils.SetupVCR(t, "datasource_node")
	config := `
resource "btpservice_ctms_node" "test" {
  name = "tf-ctms-ds-node"
}

data "btpservice_ctms_node" "id" {
  id = btpservice_ctms_node.test.id
}

data "btpservice_ctms_node" "name" {
  name       = btpservice_ctms_node.test.name
  depends_on = [btpservice_ctms_node.test]
}
`

	resource.Test(
		t,
		rec.TestCase(
			resource.TestStep{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.btpservice_ctms_node.id", "id", "btpservice_ctms_node.test", "id"),
					resource.TestCheckResourceAttrPair("data.btpservice_ctms_node.name", "id", "btpservice_ctms_node.test", "id"),
				),
			},
			resource.TestStep{
				Config:   config,
				PlanOnly: true,
			},
		),
	)
}

func TestDataSourceCTMSNodeMissing(t *testing.T) {
	rec := testutils.SetupVCR(t, "datasource_node_missing")

	resource.Test(
		t,
		rec.TestCase(
			resource.TestStep{
				Config: `data "btpservice_ctms_node" "missing" {
  name = "tf-ctms-deliberately-missing-object"
}`,
				ExpectError: regexp.MustCompile("Error Reading CTMS"),
			},
		),
	)
}
