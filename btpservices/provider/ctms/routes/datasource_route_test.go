package ctmsroutes_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/testutils"
)

func TestDataSourceCTMSRoute(t *testing.T) {
	rec := testutils.SetupVCR(t, "datasource_route")
	config := testDataSourceCTMSRouteConfig()
	var checks []statecheck.StateCheck
	for _, address := range []string{"data.btpservice_ctms_route.id", "data.btpservice_ctms_route.name"} {
		for attribute, value := range testDevelopmentToQualityRouteValues() {
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

func TestDataSourceCTMSRouteMissing(t *testing.T) {
	rec := testutils.SetupVCR(t, "datasource_route_missing")

	resource.Test(
		t,
		rec.TestCase(
			resource.TestStep{
				Config:      testDataSourceCTMSRouteMissingConfig(),
				ExpectError: regexp.MustCompile("(?s)Error Reading CTMS Route.*expected exactly one CTMS route.*found 0"),
			},
		),
	)
}

func TestDataSourceCTMSRouteValidation(t *testing.T) {
	for _, tc := range []struct{ name, attributes, errorPattern string }{
		{"missing selector", "", `Invalid Attribute Combination`},
		{"both selectors", `id = "192"
name = "Terraform_DEV_TO_TEST"`, `Invalid Attribute Combination`},
		{"empty name", `name = ""`, `Attribute name.*length must be at least 1`},
		{"empty id", `id = ""`, `Invalid CTMS ID`},
		{"non decimal id", `id = "abc"`, `Invalid CTMS ID`},
		{"leading zero id", `id = "01"`, `Invalid CTMS ID`},
		{"overflow id", `id = "9223372036854775808"`, `Invalid CTMS ID`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, routeValidationTestCase(t, resource.TestStep{
				Config:      testDataSourceCTMSRouteAttributesConfig(tc.attributes),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("(?s)" + tc.errorPattern),
			}))
		})
	}
}

func testDataSourceCTMSRouteConfig() string {
	return `
data "btpservice_ctms_route" "id" {
  id = "192"
}

data "btpservice_ctms_route" "name" {
  name = "Terraform_DEV_TO_TEST"
}
`
}

func testDataSourceCTMSRouteMissingConfig() string {
	return `data "btpservice_ctms_route" "missing" {
  name = "tf-ctms-deliberately-missing-object"
}`
}

func testDataSourceCTMSRouteAttributesConfig(attributes string) string {
	return fmt.Sprintf(`data "btpservice_ctms_route" "test" {
%s
}
`, attributes)
}
