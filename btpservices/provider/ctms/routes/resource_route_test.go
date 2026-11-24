package ctmsroutes_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/testutils"
)

func TestResourceCTMSRoute(t *testing.T) {
	rec := testutils.SetupVCR(t, "resource_route")
	routeID := statecheck.CompareValue(compare.ValuesSame())
	resource.Test(
		t,
		rec.TestCase(
			resource.TestStep{
				Config: testResourceCTMSRouteConfig("tf-ctms-route", "tf-ctms-route-source", "My Description"),
				Check:  testResourceCTMSRouteCheck("tf-ctms-route", "My Description"),
				ConfigStateChecks: []statecheck.StateCheck{
					routeID.AddStateValue("btpservice_ctms_route.test", tfjsonpath.New("id")),
				},
			},
			resource.TestStep{
				Config: testResourceCTMSRouteConfig("tf-ctms-route-renamed", "tf-ctms-route-source", "My Description Updated"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("btpservice_ctms_route.test", plancheck.ResourceActionUpdate)},
				},
				Check: testResourceCTMSRouteCheck("tf-ctms-route-renamed", "My Description Updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					routeID.AddStateValue("btpservice_ctms_route.test", tfjsonpath.New("id")),
				},
			},
			resource.TestStep{
				ResourceName:      "btpservice_ctms_route.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			resource.TestStep{
				ResourceName:    "btpservice_ctms_route.test",
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
			},
		),
	)
}

func TestResourceCTMSRouteMinimal(t *testing.T) {
	rec := testutils.SetupVCR(t, "resource_route_minimal")
	config := testResourceCTMSRouteMinimalConfig()
	resource.Test(t, rec.TestCase(
		resource.TestStep{
			Config: config,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttrSet("btpservice_ctms_route.test", "id"),
				resource.TestCheckResourceAttr("btpservice_ctms_route.test", "name", "tf-ctms-route-minimal"),
				resource.TestCheckResourceAttrPair("btpservice_ctms_route.test", "source_node_id", "btpservice_ctms_node.source", "id"),
				resource.TestCheckResourceAttrPair("btpservice_ctms_route.test", "target_node_id", "btpservice_ctms_node.target", "id"),
			),
		},
		resource.TestStep{
			ResourceName:      "btpservice_ctms_route.test",
			ImportState:       true,
			ImportStateVerify: true,
		},
		resource.TestStep{
			ResourceName:    "btpservice_ctms_route.test",
			ImportState:     true,
			ImportStateKind: resource.ImportBlockWithResourceIdentity,
		},
		resource.TestStep{Config: config, PlanOnly: true},
	))
}

func TestResourceCTMSRouteValidation(t *testing.T) {
	for _, tc := range []struct{ name, attributes, errorPattern string }{
		{"missing name", `source_node_id = "183"
target_node_id = "184"`, `The argument "name" is required`},
		{"empty name", `name = ""
source_node_id = "183"
target_node_id = "184"`, `Attribute name.*length must be at least 1`},
		{"missing source", `name = "validation"
target_node_id = "184"`, `The argument "source_node_id" is required`},
		{"missing target", `name = "validation"
source_node_id = "183"`, `The argument "target_node_id" is required`},
		{"computed id", `name = "validation"
source_node_id = "183"
target_node_id = "184"
id = "192"`, `Invalid Configuration for Read-Only Attribute`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, routeValidationTestCase(t, resource.TestStep{
				Config:      testResourceCTMSRouteAttributesConfig(tc.attributes),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("(?s)" + tc.errorPattern),
			}))
		})
	}
	for _, attribute := range []string{"source_node_id", "target_node_id"} {
		for _, tc := range []struct{ name, value string }{
			{"empty", ""},
			{"non decimal", "abc"},
			{"leading zero", "01"},
			{"overflow", "9223372036854775808"},
		} {
			t.Run(attribute+"/"+tc.name, func(t *testing.T) {
				source, target := "183", "184"
				if attribute == "source_node_id" {
					source = tc.value
				} else {
					target = tc.value
				}
				resource.Test(t, routeValidationTestCase(t, resource.TestStep{
					Config: testResourceCTMSRouteAttributesConfig(fmt.Sprintf(`name = "validation"
source_node_id = %q
target_node_id = %q`, source, target)),
					PlanOnly:    true,
					ExpectError: regexp.MustCompile("Invalid CTMS ID"),
				}))
			})
		}
	}
}

func TestResourceCTMSRouteValidConfiguration(t *testing.T) {
	for _, tc := range []struct{ name, attributes string }{
		{"required fields", ""},
		{"description", `description = "Route from Development to Test Space"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, routeValidationTestCase(t, resource.TestStep{
				Config: testResourceCTMSRouteAttributesConfig(`name = "validation"
source_node_id = "183"
target_node_id = "184"
` + tc.attributes),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			}))
		})
	}
}

func testResourceCTMSRouteCheck(name, description string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttrSet("btpservice_ctms_route.test", "id"),
		resource.TestCheckResourceAttr("btpservice_ctms_route.test", "name", name),
		resource.TestCheckResourceAttr("btpservice_ctms_route.test", "description", description),
		resource.TestCheckResourceAttrPair("btpservice_ctms_route.test", "source_node_id", "btpservice_ctms_node.source", "id"),
		resource.TestCheckResourceAttrPair("btpservice_ctms_route.test", "target_node_id", "btpservice_ctms_node.target", "id"),
	)
}

func testResourceCTMSRouteConfig(routeName, sourceName, description string) string {
	return fmt.Sprintf(`
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
`, sourceName, routeName, description)
}

func testResourceCTMSRouteMinimalConfig() string {
	return `
resource "btpservice_ctms_node" "source" {
  name = "tf-ctms-route-minimal-source"
  virtual = true
}

resource "btpservice_ctms_node" "target" {
  name = "tf-ctms-route-minimal-target"
  virtual = true
}

resource "btpservice_ctms_route" "test" {
  name = "tf-ctms-route-minimal"
  source_node_id = btpservice_ctms_node.source.id
  target_node_id = btpservice_ctms_node.target.id
}
`
}

func testResourceCTMSRouteAttributesConfig(attributes string) string {
	return fmt.Sprintf(`resource "btpservice_ctms_route" "test" {
%s
}
`, attributes)
}
