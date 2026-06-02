// btpservices/provider/cicd/repositories/datasource_repository_jobs_test.go

package cicdrepositories_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/cicd/utils"
	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/tfutils"
)

func TestDatasourceCicdRepositoryJobs(t *testing.T) {
	t.Parallel()

	t.Run("list jobs for repository", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/datasource_repository_jobs_list")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
data "btpservice_cicd_repository_jobs" "uut" {
  repository = "tf-ds-test-repo"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.btpservice_cicd_repository_jobs.uut", "repository", "tf-ds-test-repo"),
						resource.TestCheckResourceAttr("data.btpservice_cicd_repository_jobs.uut", "values.#", "1"),
						resource.TestCheckResourceAttr("data.btpservice_cicd_repository_jobs.uut", "values.0.name", "tf-ds-test-job-credentials"),
						resource.TestCheckResourceAttr("data.btpservice_cicd_repository_jobs.uut", "values.0.pipeline", "cf-env"),
						resource.TestCheckResourceAttr("data.btpservice_cicd_repository_jobs.uut", "values.0.pipeline_version", "3.0"),
						resource.TestCheckResourceAttr("data.btpservice_cicd_repository_jobs.uut", "values.0.branch", "main"),
						resource.TestCheckResourceAttr("data.btpservice_cicd_repository_jobs.uut", "values.0.active", "true"),
						resource.TestCheckResourceAttr("data.btpservice_cicd_repository_jobs.uut", "values.0.build_retention_days", "7"),
						resource.TestCheckResourceAttr("data.btpservice_cicd_repository_jobs.uut", "values.0.max_builds_to_keep", "50"),
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_repository_jobs.uut", "values.0.id"),
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_repository_jobs.uut", "values.0.repository_id"),
					),
				},
			},
		})
	})

	t.Run("repository not found", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/datasource_repository_jobs_not_found")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
data "btpservice_cicd_repository_jobs" "uut" {
  repository = "this-repository-does-not-exist"
}
`,
					ExpectError: regexp.MustCompile(`Repository Not Found`),
				},
			},
		})
	})
}
