package ctmsdestinations_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/testutils"
)

func TestDataSourceCTMSDestinations(t *testing.T) {
	rec := testutils.SetupVCR(t, "datasource_destinations")
	config := `
data "btpservice_ctms_destinations" "all" {}
`

	resource.Test(
		t,
		rec.TestCase(
			resource.TestStep{
				Config: config,
				Check:  resource.TestCheckResourceAttrSet("data.btpservice_ctms_destinations.all", "values.#"),
			},
		),
	)
}
