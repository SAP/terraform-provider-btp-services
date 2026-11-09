package ctmsnodes_test

import (
	"context"
	"fmt"
	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/testutils"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	ctmsmodels "github.com/SAP/terraform-provider-btp-services/internal/ctms/models"
)

func TestResourceCTMSNode(t *testing.T) {
	rec := testutils.SetupVCR(t, "resource_node")
	config := func(description string, upload bool) string {
		return fmt.Sprintf(
			`
resource "btpservice_ctms_node" "test" {
 name = "tf-ctms-node-lifecycle"
 description = %q
 upload_allowed = %t
 forward_mode = "MANUAL"
 import_disabled = true
 import_disabled_reason = "Terraform recording"
 auto_import = false
 virtual = false
}
`,
			description,
			upload,
		)
	}
	resource.Test(
		t,
		rec.TestCase(
			resource.TestStep{
				Config: config("initial", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("btpservice_ctms_node.test", "id"),
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "upload_allowed", "true"),
				),
			},
			resource.TestStep{
				Config: config("", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "description", ""),
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "upload_allowed", "false"),
				),
			},
			resource.TestStep{
				ResourceName:      "btpservice_ctms_node.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			resource.TestStep{
				ResourceName:      "btpservice_ctms_node.test",
				ImportState:       true,
				ImportStateKind:   resource.ImportBlockWithResourceIdentity,
				ImportStateVerify: true,
			},
			resource.TestStep{
				Config:   config("", false),
				PlanOnly: true,
			},
		),
	)
}

func TestResourceCTMSNodeMinimal(t *testing.T) {
	rec := testutils.SetupVCR(t, "resource_node_minimal")
	config := `resource "btpservice_ctms_node" "test" { name = "tf-ctms-node-minimal" }`
	resource.Test(
		t,
		rec.TestCase(
			resource.TestStep{Config: config},
			resource.TestStep{
				Config:   config,
				PlanOnly: true,
			},
			resource.TestStep{
				ResourceName:      "btpservice_ctms_node.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		),
	)
}

func TestResourceCTMSNodeTargets(t *testing.T) {
	rec := testutils.SetupVCR(t, "resource_node_targets", "test_destination")
	base := `
resource "btpservice_ctms_node" "test" {
 name = "tf-ctms-node-targets"
 import_disabled = true
 auto_import = false
 %s
}
`
	target := func(strategy string) string {
		return fmt.Sprintf(
			`targets = [{ content_type = "MTA", destination_name = %q, import_options = { strategy = %q } }]`,
			rec.Values["test_destination"],
			strategy,
		)
	}
	resource.Test(
		t,
		rec.TestCase(
			resource.TestStep{
				Config: fmt.Sprintf(base, target("default")),
				Check:  resource.TestCheckResourceAttr("btpservice_ctms_node.test", "targets.0.import_options.strategy", "default"),
			},
			resource.TestStep{
				Config: fmt.Sprintf(base, target("blue-green")),
				Check:  resource.TestCheckResourceAttr("btpservice_ctms_node.test", "targets.0.import_options.strategy", "blue-green"),
			},
			resource.TestStep{
				ResourceName:      "btpservice_ctms_node.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// This is a live-contract check, not a claim that the API accepts every reset.
			resource.TestStep{
				Config: fmt.Sprintf(base, "targets = []"),
				Check:  resource.TestCheckResourceAttr("btpservice_ctms_node.test", "targets.#", "0"),
			},
			resource.TestStep{
				Config:   fmt.Sprintf(base, "targets = []"),
				PlanOnly: true,
			},
		),
	)
}

func TestResourceCTMSNodeVirtual(t *testing.T) {
	rec := testutils.SetupVCR(t, "resource_node_virtual")
	config := `
resource "btpservice_ctms_node" "test" {
 name = "tf-ctms-node-virtual"
 virtual = true
 import_disabled = true
 auto_import = true
}
`
	resource.Test(
		t,
		rec.TestCase(
			resource.TestStep{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "virtual", "true"),
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "auto_import", "true"),
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "import_disabled", "true"),
				),
			},
			resource.TestStep{
				ResourceName:      "btpservice_ctms_node.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			resource.TestStep{
				Config:   config,
				PlanOnly: true,
			},
		),
	)
}

func TestResourceCTMSNodeDriftAndDeletion(t *testing.T) {
	rec := testutils.SetupVCR(t, "resource_node_drift")
	var id int64
	capture := func(state *terraform.State) error {
		item, ok := state.RootModule().Resources["btpservice_ctms_node.test"]
		if !ok {
			return fmt.Errorf("node missing from state")
		}
		var err error
		id, err = strconv.ParseInt(item.Primary.ID, 10, 64)
		return err
	}
	managed := `resource "btpservice_ctms_node" "test" {
 name = "tf-ctms-node-drift"
 description = "managed"
}`
	unmanaged := `resource "btpservice_ctms_node" "test" { name = "tf-ctms-node-drift" }`
	change := func() {
		t.Helper()
		description := "external"
		if _, err := rec.Client("terraform-provider-ctms-test").Nodes.Update(context.Background(), id, ctmsmodels.NodeUpdateRequest{Description: &description}); err != nil {
			t.Fatal(err)
		}
	}
	remove := func() {
		t.Helper()
		if err := rec.Client("terraform-provider-ctms-test").Nodes.Delete(context.Background(), id); err != nil {
			t.Fatal(err)
		}
	}
	resource.Test(
		t,
		rec.TestCase(
			resource.TestStep{
				Config: managed,
				Check:  capture,
			},
			resource.TestStep{
				PreConfig: change,
				Config:    managed,
				Check:     resource.TestCheckResourceAttr("btpservice_ctms_node.test", "description", "managed"),
			},
			resource.TestStep{Config: unmanaged},
			resource.TestStep{
				PreConfig: change,
				Config:    unmanaged,
				PlanOnly:  true,
			},
			resource.TestStep{
				PreConfig: remove,
				Config:    unmanaged,
				Check:     capture,
			},
			resource.TestStep{
				ResourceName:      "btpservice_ctms_node.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		),
	)
}
