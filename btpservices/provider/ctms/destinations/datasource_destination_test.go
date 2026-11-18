package ctmsdestinations_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/testutils"
)

func TestDataSourceCTMSDestination(t *testing.T) {
	rec := testutils.SetupVCR(t, "datasource_destination")
	config := testDataSourceCTMSDestinationConfig()
	var checks []statecheck.StateCheck
	for attribute, value := range testDestinationSummaryValues("ctms-deployer-dev") {
		checks = append(checks, statecheck.ExpectKnownValue("data.btpservice_ctms_destination.test", tfjsonpath.New(attribute), value))
	}

	resource.Test(
		t,
		rec.TestCase(
			resource.TestStep{
				Config:            config,
				Check:             resource.TestCheckResourceAttr("data.btpservice_ctms_destination.test", "include_tenant_id", "false"),
				ConfigStateChecks: checks,
			},
			resource.TestStep{
				Config:   config,
				PlanOnly: true,
			},
		),
	)
}

func testDataSourceCTMSDestinationConfig() string {
	return `
data "btpservice_ctms_destination" "test" {
  name              = "ctms-deployer-dev"
  include_tenant_id = false
}
`
}

func TestDataSourceCTMSDestinationMissing(t *testing.T) {
	rec := testutils.SetupVCR(t, "datasource_destination_missing")

	resource.Test(
		t,
		rec.TestCase(
			resource.TestStep{
				Config: `
data "btpservice_ctms_destination" "missing" {
  name = "tf-ctms-deliberately-missing-object"
}
`,
				// Require a 404 for the requested destination path.
				ExpectError: regexp.MustCompile(
					`(?s)Error Reading CTMS Destination.*CTMS API error 404 for\s+` +
						`/v2/destinations/tf-ctms-deliberately-missing-object\b`,
				),
			},
		),
	)
}
