// btpservices/provider/cicd/jobs/datasource_triggers_test.go

package cicdjobs_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/cicd/utils"
	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/tfutils"
)

func TestDatasourceCicdTriggers(t *testing.T) {
	t.Parallel()

	t.Run("list", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/datasource_triggers_list")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
data "btpservice_cicd_triggers" "uut" {
  job = "tf-test-job"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_triggers.uut", "id"),
						resource.TestCheckResourceAttr("data.btpservice_cicd_triggers.uut", "job", "tf-test-job"),
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_triggers.uut", "values.#"),
						resource.TestCheckResourceAttr("data.btpservice_cicd_triggers.uut", "values.0.id", "6a9c2374-1e61-4cc9-9056-28a90f4ab3bb"),
						resource.TestCheckResourceAttr("data.btpservice_cicd_triggers.uut", "values.0.type", "timer"),
						resource.TestCheckResourceAttr("data.btpservice_cicd_triggers.uut", "values.0.timer.branch", "main"),
						resource.TestCheckResourceAttr("data.btpservice_cicd_triggers.uut", "values.0.timer.cron", "0 9 * * 1-5"),
					),
				},
			},
		})
	})
}
