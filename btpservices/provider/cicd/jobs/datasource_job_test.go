// btpservices/provider/cicd/jobs/datasource_job_test.go

package cicdjobs_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/cicd/utils"
	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/tfutils"
)

func TestDatasourceCicdJob(t *testing.T) {
	t.Parallel()

	t.Run("read job by name", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/datasource_job_by_name")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
data "btpservice_cicd_job" "uut" {
  name = "tf-test-job"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_job.uut", "id"),
						resource.TestCheckResourceAttr("data.btpservice_cicd_job.uut", "name", "tf-test-job"),
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_job.uut", "pipeline"),
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_job.uut", "pipeline_version"),
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_job.uut", "pipeline_parameters"),
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_job.uut", "repository_id"),
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_job.uut", "branch"),
					),
				},
			},
		})
	})

	t.Run("read job by id", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/datasource_job_by_id")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
data "btpservice_cicd_job" "uut" {
  id = "ca9101f6-49b7-4e91-b632-e4c222fd9e4d"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.btpservice_cicd_job.uut", "id", "ca9101f6-49b7-4e91-b632-e4c222fd9e4d"),
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_job.uut", "name"),
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_job.uut", "pipeline"),
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_job.uut", "pipeline_parameters"),
					),
				},
			},
		})
	})

	t.Run("job not found", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/datasource_job_not_found")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
data "btpservice_cicd_job" "uut" {
  name = "this-job-does-not-exist"
}
`,
					ExpectError: regexp.MustCompile(`Job Not Found`),
				},
			},
		})
	})

	t.Run("error - neither id nor name set", func(t *testing.T) {
		t.Parallel()
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(utils.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(utils.Redacted) + `
data "btpservice_cicd_job" "uut" {}
`,
					ExpectError: regexp.MustCompile(`(?i)Invalid Attribute Combination`),
				},
			},
		})
	})

	t.Run("error - both id and name set", func(t *testing.T) {
		t.Parallel()
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(utils.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(utils.Redacted) + `
data "btpservice_cicd_job" "uut" {
  id   = "some-id"
  name = "some-name"
}
`,
					ExpectError: regexp.MustCompile(`(?i)Invalid Attribute Combination`),
				},
			},
		})
	})
}
