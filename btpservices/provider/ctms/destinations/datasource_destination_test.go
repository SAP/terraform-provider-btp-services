package ctmsdestinations_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/testutils"
)

func TestDataSourceCTMSDestination(t *testing.T) {
	rec := testutils.SetupVCR(t, "datasource_destination", "test_destination")
	config := func(include bool) string {
		return fmt.Sprintf(
			`
data "btpservice_ctms_destination" "test" {
  name              = %q
  include_tenant_id = %t
}
`,
			rec.Values["test_destination"],
			include,
		)
	}

	resource.Test(
		t,
		rec.TestCase(
			resource.TestStep{
				Config: config(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.btpservice_ctms_destination.test", "name", rec.Values["test_destination"]),
					resource.TestCheckResourceAttr("data.btpservice_ctms_destination.test", "include_tenant_id", "false"),
				),
			},
			resource.TestStep{
				Config: config(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.btpservice_ctms_destination.test", "name", rec.Values["test_destination"]),
					resource.TestCheckResourceAttr("data.btpservice_ctms_destination.test", "include_tenant_id", "true"),
				),
			},
		),
	)
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
