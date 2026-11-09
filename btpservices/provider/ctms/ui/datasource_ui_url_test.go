package ctmsui_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/testutils"
)

func TestDataSourceCTMSUIURL(t *testing.T) {
	rec := testutils.SetupVCR(t, "datasource_ui_url")

	resource.Test(
		t,
		rec.TestCase(
			resource.TestStep{
				Config: `data "btpservice_ctms_ui_url" "test" {}`,
				Check:  resource.TestCheckResourceAttrSet("data.btpservice_ctms_ui_url.test", "url"),
			},
		),
	)
}
