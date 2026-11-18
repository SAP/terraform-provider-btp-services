package ctmsdestinations_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/testutils"
)

func TestDataSourceCTMSDestinations(t *testing.T) {
	rec := testutils.SetupVCR(t, "datasource_destinations")
	config := testDataSourceCTMSDestinationsConfig()

	resource.Test(
		t,
		rec.TestCase(
			resource.TestStep{
				Config: config,
				Check:  resource.TestCheckResourceAttrSet("data.btpservice_ctms_destinations.all", "values.#"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.btpservice_ctms_destinations.all", tfjsonpath.New("values"), knownvalue.SetPartial([]knownvalue.Check{
						knownvalue.ObjectExact(testDestinationSummaryValues("ctms-deployer-dev")),
						knownvalue.ObjectExact(testDestinationSummaryValues("ctms-deployer-test")),
						knownvalue.ObjectExact(testDestinationSummaryValues("ctms-deployer-prod")),
					})),
				},
			},
			resource.TestStep{
				Config:   config,
				PlanOnly: true,
			},
		),
	)
}

func testDataSourceCTMSDestinationsConfig() string {
	return `data "btpservice_ctms_destinations" "all" {}`
}
