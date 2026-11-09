package ctmsroutes_test

import (
	"fmt"
	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestResourceCTMSRoute(t *testing.T) {
	rec := testutils.SetupVCR(t, "resource_route")
	config := func(routeName, sourceName, description string) string {
		return fmt.Sprintf(
			`
resource "btpservice_ctms_node" "source" {
 name = %q
 virtual = true
 lifecycle { create_before_destroy = true }
}
resource "btpservice_ctms_node" "target" {
 name = "tf-ctms-route-target"
 virtual = true
}
resource "btpservice_ctms_route" "test" {
 name = %q
 description = %q
 source_node_id = btpservice_ctms_node.source.id
 target_node_id = btpservice_ctms_node.target.id
}
`,
			sourceName,
			routeName,
			description,
		)
	}
	resource.Test(
		t,
		rec.TestCase(
			resource.TestStep{
				Config: config("tf-ctms-route", "tf-ctms-route-source", "initial"),
				Check:  resource.TestCheckResourceAttrSet("btpservice_ctms_route.test", "id"),
			},
			resource.TestStep{
				Config: config("tf-ctms-route-renamed", "tf-ctms-route-source", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("btpservice_ctms_route.test", "name", "tf-ctms-route-renamed"),
					resource.TestCheckResourceAttr("btpservice_ctms_route.test", "description", ""),
				),
			},
			// Replacement of the source must rewire the existing route before removing
			// the old node. A second incoming route must not be created.
			resource.TestStep{
				Config: config("tf-ctms-route-renamed", "tf-ctms-route-source-new", ""),
				Check:  resource.TestCheckResourceAttrPair("btpservice_ctms_route.test", "source_node_id", "btpservice_ctms_node.source", "id"),
			},
			resource.TestStep{
				ResourceName:      "btpservice_ctms_route.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			resource.TestStep{
				ResourceName:      "btpservice_ctms_route.test",
				ImportState:       true,
				ImportStateKind:   resource.ImportBlockWithResourceIdentity,
				ImportStateVerify: true,
			},
			resource.TestStep{
				Config:   config("tf-ctms-route-renamed", "tf-ctms-route-source-new", ""),
				PlanOnly: true,
			},
		),
	)
}
