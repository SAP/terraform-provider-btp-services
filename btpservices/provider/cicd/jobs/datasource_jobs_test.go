// btpservices/provider/cicd/jobs/datasource_jobs_test.go

package cicdjobs_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/cicd/utils"
	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/tfutils"
)

func TestDatasourceCicdJobs(t *testing.T) {
	t.Parallel()

	t.Run("list all jobs", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/datasource_jobs_list")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
data "btpservice_cicd_jobs" "uut" {}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_jobs.uut", "values.#"),
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_jobs.uut", "values.0.id"),
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_jobs.uut", "values.0.name"),
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_jobs.uut", "values.0.pipeline"),
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_jobs.uut", "values.0.pipeline_version"),
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_jobs.uut", "values.0.pipeline_parameters"),
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_jobs.uut", "values.0.repository_id"),
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_jobs.uut", "values.0.branch"),
					),
				},
			},
		})
	})
}
