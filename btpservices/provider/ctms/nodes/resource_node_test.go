package ctmsnodes_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/testutils"
)

func TestResourceCTMSNode(t *testing.T) {
	rec := testutils.SetupVCR(t, "resource_node")
	resource.Test(
		t,
		rec.TestCase(
			resource.TestStep{
				Config: testResourceCTMSNodeConfig("Development Space", true, "AUTO"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("btpservice_ctms_node.test", "id"),
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "upload_allowed", "true"),
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "name", "tf-ctms-node-lifecycle"),
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "description", "Development Space"),
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "forward_mode", "AUTO"),
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "auto_import", "false"),
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "virtual", "false"),
				),
			},

			resource.TestStep{
				Config: testResourceCTMSNodeConfig("Development Space Updated", false, "MANUAL"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("btpservice_ctms_node.test", plancheck.ResourceActionUpdate)},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "upload_allowed", "false"),
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "name", "tf-ctms-node-lifecycle"),
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "description", "Development Space Updated"),
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "forward_mode", "MANUAL"),
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "auto_import", "false"),
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "virtual", "false"),
				),
			},

			resource.TestStep{
				ResourceName:      "btpservice_ctms_node.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			resource.TestStep{
				ResourceName:    "btpservice_ctms_node.test",
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
			},
		),
	)
}

func TestResourceCTMSNode2(t *testing.T) {
	rec := testutils.SetupVCR(t, "resource_node2")
	resource.Test(
		t,
		rec.TestCase(
			resource.TestStep{
				Config: testResourceCTMSNodeConfig("", true, "AUTO"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("btpservice_ctms_node.test", "id"),
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "upload_allowed", "true"),
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "name", "tf-ctms-node-lifecycle"),
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "description", ""),
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "forward_mode", "AUTO"),
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "auto_import", "false"),
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "virtual", "false"),
				),
			},

			resource.TestStep{
				Config: testResourceCTMSNodeConfig("Development Space Updated", false, "MANUAL"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("btpservice_ctms_node.test", plancheck.ResourceActionUpdate)},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "upload_allowed", "false"),
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "name", "tf-ctms-node-lifecycle"),
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "description", "Development Space Updated"),
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "forward_mode", "MANUAL"),
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "auto_import", "false"),
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "virtual", "false"),
				),
			},

			resource.TestStep{
				ResourceName:      "btpservice_ctms_node.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			resource.TestStep{
				ResourceName:    "btpservice_ctms_node.test",
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
			},
		),
	)
}
func TestResourceCTMSNodeMinimal(t *testing.T) {
	rec := testutils.SetupVCR(t, "resource_node_minimal")
	config := testResourceCTMSNodeMinimalConfig("tf-ctms-node-minimal")
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
			resource.TestStep{
				Config: testResourceCTMSNodeMinimalConfig("tf-ctms-node-minimal-renamed"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("btpservice_ctms_node.test", plancheck.ResourceActionDestroyBeforeCreate)},
				},
				Check: resource.TestCheckResourceAttr("btpservice_ctms_node.test", "name", "tf-ctms-node-minimal-renamed"),
			},
		),
	)
}

func TestResourceCTMSNodeTargets(t *testing.T) {
	rec := testutils.SetupVCR(t, "resource_node_targets")
	resource.Test(
		t,
		rec.TestCase(
			resource.TestStep{
				Config: testResourceCTMSNodeTargetsConfig("ctms-deployer-dev", "default"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "targets.#", "1"),
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "targets.0.content_type", "MTA"),
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "targets.0.destination_name", "ctms-deployer-dev"),
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "targets.0.import_options.strategy", "default"),
				),
			},
			resource.TestStep{
				Config: testResourceCTMSNodeTargetsConfig("ctms-deployer-dev", "blue-green"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "targets.#", "1"),
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "targets.0.content_type", "MTA"),
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "targets.0.destination_name", "ctms-deployer-dev"),
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "targets.0.import_options.strategy", "blue-green"),
				),
			},
			resource.TestStep{
				ResourceName:      "btpservice_ctms_node.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// This is a live-contract check, not a claim that the API accepts every reset.
			resource.TestStep{
				Config: testResourceCTMSNodeTargetsConfig("", ""),
				Check:  resource.TestCheckResourceAttr("btpservice_ctms_node.test", "targets.#", "0"),
			},
			resource.TestStep{
				Config:   testResourceCTMSNodeTargetsConfig("", ""),
				PlanOnly: true,
			},
		),
	)
}

func TestResourceCTMSNodeVirtual(t *testing.T) {
	rec := testutils.SetupVCR(t, "resource_node_virtual")
	config := testResourceCTMSNodeVirtualConfig()
	resource.Test(
		t,
		rec.TestCase(
			resource.TestStep{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "virtual", "true"),
					resource.TestCheckResourceAttr("btpservice_ctms_node.test", "auto_import", "true"),
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

func TestResourceCTMSNodeValidation(t *testing.T) {
	for _, tc := range []struct{ name, attributes, errorPattern string }{
		{"missing name", "", `The argument "name" is required`},
		{"empty name", `name = ""`, `Attribute name.*length must be at least 1`},
		{"invalid forward mode", `name = "validation"
forward_mode = "INVALID"`, `Attribute forward_mode.*one of`},
		{"missing content type", `name = "validation"
targets = [{ destination_name = "destination" }]`, `"content_type" is required`},
		{"missing destination", `name = "validation"
targets = [{ content_type = "MTA" }]`, `"destination_name" is required`},
		{"empty destination", `name = "validation"
targets = [{ content_type = "MTA", destination_name = "" }]`, `destination_name.*length must be at least 1`},
		{"invalid content type", `name = "validation"
targets = [{ content_type = "INVALID", destination_name = "destination" }]`, `content_type.*one of`},
		{"invalid strategy", `name = "validation"
targets = [{ content_type = "MTA", destination_name = "destination", import_options = { strategy = "INVALID" } }]`, `strategy.*one of`},
		{"virtual node with targets", `name = "validation"
virtual = true
targets = [{ content_type = "MTA", destination_name = "destination" }]`, `Virtual Node Has Targets`},
		{"computed id", `name = "validation"
id = "123"`, `Invalid Configuration for Read-Only Attribute`},
		{"computed controller", `name = "validation"
controlled_by = "application"`, `Invalid Configuration for Read-Only Attribute`},
		{"computed job", `name = "validation"
job_id = 123`, `Invalid Configuration for Read-Only Attribute`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, nodeValidationTestCase(t, resource.TestStep{
				Config:      testResourceCTMSNodeAttributesConfig(tc.attributes),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("(?s)" + tc.errorPattern),
			}))
		})
	}
	for _, contentType := range []string{"XSCDU", "APP", "BTP_ABAP"} {
		t.Run("strategy requires MTA/"+contentType, func(t *testing.T) {
			resource.Test(t, nodeValidationTestCase(t, resource.TestStep{
				Config: testResourceCTMSNodeAttributesConfig(fmt.Sprintf(`name = "validation"
targets = [{ content_type = %q, destination_name = "destination", import_options = { strategy = "default" } }]`, contentType)),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("Invalid Deployment Strategy"),
			}))
		})
	}
}

func TestResourceCTMSNodeValidConfiguration(t *testing.T) {
	for _, mode := range []string{"AUTO", "POST_IMPORT", "ON_SUCCESS", "MANUAL"} {
		t.Run("forward mode/"+mode, func(t *testing.T) {
			resource.Test(t, nodeValidationTestCase(t, resource.TestStep{
				Config:             testResourceCTMSNodeConfig("Development Space", true, mode),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			}))
		})
	}
	for _, contentType := range []string{"MTA", "XSCDU", "APP", "BTP_ABAP"} {
		t.Run("content type/"+contentType, func(t *testing.T) {
			resource.Test(t, nodeValidationTestCase(t, resource.TestStep{
				Config: testResourceCTMSNodeAttributesConfig(fmt.Sprintf(`name = "validation"
targets = [{ content_type = %q, destination_name = "destination" }]`, contentType)),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			}))
		})
	}
	for _, strategy := range []string{"default", "blue-green"} {
		t.Run("strategy/"+strategy, func(t *testing.T) {
			resource.Test(t, nodeValidationTestCase(t, resource.TestStep{
				Config:             testResourceCTMSNodeTargetsConfig("destination", strategy),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			}))
		})
	}
	t.Run("virtual node without targets", func(t *testing.T) {
		resource.Test(t, nodeValidationTestCase(t, resource.TestStep{
			Config:             testResourceCTMSNodeVirtualConfig(),
			PlanOnly:           true,
			ExpectNonEmptyPlan: true,
		}))
	})
}

func testResourceCTMSNodeConfig(description string, upload bool, forwardMode string) string {
	return fmt.Sprintf(
		`
resource "btpservice_ctms_node" "test" {
 name = "tf-ctms-node-lifecycle"
 description = %q
 upload_allowed = %t
 forward_mode = %q
 auto_import = false
 virtual = false
}
`,
		description,
		upload,
		forwardMode,
	)
}

func testResourceCTMSNodeMinimalConfig(name string) string {
	return testResourceCTMSNodeAttributesConfig(fmt.Sprintf("name = %q", name))
}

func testResourceCTMSNodeVirtualConfig() string {
	return `
resource "btpservice_ctms_node" "test" {
 name = "tf-ctms-node-virtual"
 virtual = true
 auto_import = true
}
`
}

func testResourceCTMSNodeTargetsConfig(destination, strategy string) string {
	base := `
resource "btpservice_ctms_node" "test" {
 name = "tf-ctms-node-targets"
 auto_import = false
 %s
}
`

	if destination == "" {
		return fmt.Sprintf(base, "targets = []")
	}
	targets := fmt.Sprintf(`targets = [{ content_type = "MTA", destination_name = %q, import_options = { strategy = %q } }]`, destination, strategy)
	return fmt.Sprintf(base, targets)
}

func testResourceCTMSNodeDriftConfig() string {
	return `resource "btpservice_ctms_node" "test" {
 name = "tf-ctms-node-drift"
 description = "managed"
}`
}

func testResourceCTMSNodeAttributesConfig(attributes string) string {
	return fmt.Sprintf("resource \"btpservice_ctms_node\" \"test\" {\n%s\n}\n", attributes)
}
