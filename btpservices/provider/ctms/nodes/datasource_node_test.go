package ctmsnodes_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/testutils"
)

func TestDataSourceCTMSNode(t *testing.T) {
	rec := testutils.SetupVCR(t, "datasource_node")
	config := testDataSourceCTMSNodeConfig()
	expected := map[string]knownvalue.Check{
		"id":                   knownvalue.StringExact("183"),
		"name":                 knownvalue.StringExact("Terraform_dev"),
		"description":          knownvalue.StringExact("Development Space"),
		"forward_mode":         knownvalue.StringExact("AUTO"),
		"upload_allowed":       knownvalue.Bool(true),
		"notification_enabled": knownvalue.Bool(true),
		"auto_import":          knownvalue.Bool(false),
		"virtual":              knownvalue.Bool(false),
		"controlled_by":        knownvalue.Null(),
		"job_id":               knownvalue.Null(),
		"targets": knownvalue.ListExact([]knownvalue.Check{
			knownvalue.ObjectExact(map[string]knownvalue.Check{
				"content_type":     knownvalue.StringExact("MTA"),
				"destination_name": knownvalue.StringExact("ctms-deployer-dev"),
				"import_options": knownvalue.ObjectExact(map[string]knownvalue.Check{
					"strategy": knownvalue.StringExact("default"),
				}),
			}),
		}),
	}
	var checks []statecheck.StateCheck
	for _, address := range []string{"data.btpservice_ctms_node.id", "data.btpservice_ctms_node.name"} {
		for attribute, value := range expected {
			checks = append(checks, statecheck.ExpectKnownValue(address, tfjsonpath.New(attribute), value))
		}
	}

	resource.Test(
		t,
		rec.TestCase(
			resource.TestStep{
				Config:            config,
				ConfigStateChecks: checks,
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
				Config:      testDataSourceCTMSNodeMissingConfig(),
				ExpectError: regexp.MustCompile("(?s)Error Reading CTMS Node.*expected exactly one CTMS node.*found 0"),
			},
		),
	)
}

func TestDataSourceCTMSNodeValidation(t *testing.T) {
	for _, tc := range []struct{ name, attributes, errorPattern string }{
		{"missing selector", "", `Invalid Attribute Combination`},
		{"both selectors", `id = "123"
name = "node"`, `Invalid Attribute Combination`},
		{"empty name", `name = ""`, `Attribute name.*length must be at least 1`},
		{"empty id", `id = ""`, `Invalid CTMS ID`},
		{"non decimal id", `id = "abc"`, `Invalid CTMS ID`},
		{"leading zero id", `id = "01"`, `Invalid CTMS ID`},
		{"overflow id", `id = "9223372036854775808"`, `Invalid CTMS ID`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, nodeValidationTestCase(t, resource.TestStep{
				Config:      testDataSourceCTMSNodeAttributesConfig(tc.attributes),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("(?s)" + tc.errorPattern),
			}))
		})
	}
}

func testDataSourceCTMSNodeAttributesConfig(attributes string) string {
	return fmt.Sprintf(`data "btpservice_ctms_node" "test" {
%s
}
`, attributes)
}

func testDataSourceCTMSNodeConfig() string {
	return `
data "btpservice_ctms_node" "id" {
  id = "183"
}

data "btpservice_ctms_node" "name" {
  name = "Terraform_dev"
}
`
}

func testDataSourceCTMSNodeMissingConfig() string {
	return `data "btpservice_ctms_node" "missing" {
  name = "tf-ctms-deliberately-missing-object"
}`
}
