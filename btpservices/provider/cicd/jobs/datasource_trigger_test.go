// btpservices/provider/cicd/jobs/datasource_trigger_test.go

package cicdjobs_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/cicd/utils"
	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/tfutils"
)

func TestDatasourceCicdTrigger(t *testing.T) {
	t.Parallel()

	t.Run("read", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/datasource_trigger_read")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
data "btpservice_cicd_trigger" "uut" {
  job = "tf-test-job"
  id  = "6a9c2374-1e61-4cc9-9056-28a90f4ab3bb"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.btpservice_cicd_trigger.uut", "job", "tf-test-job"),
						resource.TestCheckResourceAttr("data.btpservice_cicd_trigger.uut", "id", "6a9c2374-1e61-4cc9-9056-28a90f4ab3bb"),
						resource.TestCheckResourceAttr("data.btpservice_cicd_trigger.uut", "type", "timer"),
						resource.TestCheckResourceAttr("data.btpservice_cicd_trigger.uut", "timer.branch", "main"),
						resource.TestCheckResourceAttr("data.btpservice_cicd_trigger.uut", "timer.cron", "0 9 * * 1-5"),
					),
				},
			},
		})
	})

	t.Run("not_found", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/datasource_trigger_not_found")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
data "btpservice_cicd_trigger" "uut" {
  job = "tf-test-job"
  id  = "00000000-0000-0000-0000-000000000000"
}
`,
					ExpectError: regexp.MustCompile(`Trigger Not Found`),
				},
			},
		})
	})
}
