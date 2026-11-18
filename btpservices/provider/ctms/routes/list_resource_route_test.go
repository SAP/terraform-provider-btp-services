package ctmsroutes_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/querycheck/queryfilter"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/testutils"
)

func TestListResourceCTMSRoute(t *testing.T) {
	rec := testutils.SetupVCR(t, "list_resource_route")
	identity := map[string]knownvalue.Check{
		"id":       knownvalue.StringExact("192"),
		"endpoint": knownvalue.StringExact(strings.TrimRight(rec.Values["endpoint"], "/")),
	}
	var checks []querycheck.KnownValueCheck
	for attribute, value := range testDevelopmentToQualityRouteValues() {
		checks = append(checks, querycheck.KnownValueCheck{
			Path: tfjsonpath.New(attribute), KnownValue: value,
		})
	}

	resource.Test(t, rec.TestCase(
		resource.TestStep{
			Query:  true,
			Config: testListResourceCTMSRouteConfig(false),
			QueryResultChecks: []querycheck.QueryResultCheck{
				querycheck.ExpectLength("btpservice_ctms_route.routes", 1),
				querycheck.ExpectIdentity("btpservice_ctms_route.routes", identity),
			},
		},
		resource.TestStep{
			Query:          true,
			GenerateConfig: true,
			Config:         testListResourceCTMSRouteConfig(true),
			QueryResultChecks: []querycheck.QueryResultCheck{
				querycheck.ExpectLength("btpservice_ctms_route.routes", 1),
				querycheck.ExpectIdentity("btpservice_ctms_route.routes", identity),
				querycheck.ExpectResourceKnownValues(
					"btpservice_ctms_route.routes", queryfilter.ByResourceIdentity(identity), checks,
				),
			},
		},
		resource.TestStep{
			Query:  true,
			Config: testListResourceCTMSRouteMissingConfig(),
			QueryResultChecks: []querycheck.QueryResultCheck{
				querycheck.ExpectLength("btpservice_ctms_route.routes", 0),
			},
		},
	))
}

func testListResourceCTMSRouteConfig(includeResource bool) string {
	return fmt.Sprintf(`
provider "btpservice" {}

list "btpservice_ctms_route" "routes" {
  provider         = btpservice
  include_resource = %t
  config {
    name           = "Terraform_DEV_TO_TEST"
    source_node_id = "183"
    target_node_id = "184"
  }
}
`, includeResource)
}

func testListResourceCTMSRouteMissingConfig() string {
	return `
provider "btpservice" {}

list "btpservice_ctms_route" "routes" {
  provider = btpservice
  config { name = "tf-ctms-deliberately-missing-route" }
}
`
}
