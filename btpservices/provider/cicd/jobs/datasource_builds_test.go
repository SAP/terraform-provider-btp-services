// btpservices/provider/cicd/jobs/datasource_builds_test.go

package cicdjobs_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/cicd/utils"
	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/tfutils"
)

func TestDatasourceCicdBuilds(t *testing.T) {
	t.Parallel()

	t.Run("returns latest build", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/datasource_builds_latest")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
data "btpservice_cicd_builds" "uut" {
  job    = "tf-test-job"
  filter = "latest"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.btpservice_cicd_builds.uut", "job", "tf-test-job"),
						resource.TestCheckResourceAttr("data.btpservice_cicd_builds.uut", "filter", "latest"),
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_builds.uut", "builds.#"),
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_builds.uut", "builds.0.id"),
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_builds.uut", "builds.0.number"),
					),
				},
			},
		})
	})

	t.Run("returns latest finished build", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/datasource_builds_latest_finished")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
data "btpservice_cicd_builds" "uut" {
  job    = "tf-test-job"
  filter = "latestFinished"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.btpservice_cicd_builds.uut", "filter", "latestFinished"),
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_builds.uut", "builds.0.id"),
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_builds.uut", "builds.0.number"),
					),
				},
			},
		})
	})

	t.Run("job not found", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/datasource_builds_job_not_found")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
data "btpservice_cicd_builds" "uut" {
  job    = "this-job-does-not-exist"
  filter = "latest"
}
`,
					ExpectError: regexp.MustCompile(`Job Not Found`),
				},
			},
		})
	})

	t.Run("error on invalid filter value", func(t *testing.T) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(utils.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(utils.Redacted) + `
data "btpservice_cicd_builds" "uut" {
  job    = "tf-test-job"
  filter = "all"
}
`,
					ExpectError: regexp.MustCompile(`(?i)Invalid Attribute Value Match`),
				},
			},
		})
	})

	t.Run("error when job attribute missing", func(t *testing.T) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(utils.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(utils.Redacted) + `
data "btpservice_cicd_builds" "uut" {
  filter = "latest"
}
`,
					ExpectError: regexp.MustCompile(`(?i)Missing required argument`),
				},
			},
		})
	})
}
