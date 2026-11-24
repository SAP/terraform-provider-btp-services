package ctmsroutes_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/testutils"
)

func TestDataSourceCTMSRoutes(t *testing.T) {
	rec := testutils.SetupVCR(t, "datasource_routes")
	config := testDataSourceCTMSRoutesConfig()

	resource.Test(
		t,
		rec.TestCase(
			resource.TestStep{
				Config: config,
				Check:  resource.TestCheckResourceAttrSet("data.btpservice_ctms_routes.all", "values.#"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.btpservice_ctms_routes.all", tfjsonpath.New("values"), knownvalue.SetPartial([]knownvalue.Check{
						knownvalue.ObjectExact(testDevelopmentToQualityRouteValues()),
						knownvalue.ObjectExact(testTestToProductionRouteValues()),
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

func testDataSourceCTMSRoutesConfig() string {
	return `data "btpservice_ctms_routes" "all" {}`
}
