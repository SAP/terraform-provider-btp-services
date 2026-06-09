// btpservices/provider/cicd/jobs/action_abort_build_test.go

package cicdjobs_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/cicd/utils"
	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/tfutils"
)

func TestActionAbortBuild(t *testing.T) {
	t.Parallel()

	t.Run("aborts build successfully", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/action_abort_build")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest: true,
			TerraformVersionChecks: []tfversion.TerraformVersionCheck{
				tfversion.SkipBelow(tfversion.Version1_14_0),
			},
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + hclActionAbortBuild("tf-test-job", 10),
				},
			},
		})
	})

	t.Run("error when build not found (404)", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/action_abort_build_not_found")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest: true,
			TerraformVersionChecks: []tfversion.TerraformVersionCheck{
				tfversion.SkipBelow(tfversion.Version1_14_0),
			},
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config:      utils.HCLProviderBlock(creds) + hclActionAbortBuild("tf-test-job", 999),
					ExpectError: regexp.MustCompile(`Build Not Found`),
				},
			},
		})
	})

	t.Run("error when job attribute missing", func(t *testing.T) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			IsUnitTest: true,
			TerraformVersionChecks: []tfversion.TerraformVersionCheck{
				tfversion.SkipBelow(tfversion.Version1_14_0),
			},
			ProtoV6ProviderFactories: utils.GetTestProviders(utils.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(utils.Redacted) + `
action "btpservice_cicd_abort_build" "uut" {
  config {
    build = 1
  }
}
`,
					ExpectError: regexp.MustCompile(`(?i)Missing required argument`),
				},
			},
		})
	})

	t.Run("error when build attribute missing", func(t *testing.T) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			IsUnitTest: true,
			TerraformVersionChecks: []tfversion.TerraformVersionCheck{
				tfversion.SkipBelow(tfversion.Version1_14_0),
			},
			ProtoV6ProviderFactories: utils.GetTestProviders(utils.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(utils.Redacted) + `
action "btpservice_cicd_abort_build" "uut" {
  config {
    job = "tf-test-job"
  }
}
`,
					ExpectError: regexp.MustCompile(`(?i)Missing required argument`),
				},
			},
		})
	})
}

func hclActionAbortBuild(job string, build int) string {
	return fmt.Sprintf(`
resource "terraform_data" "trigger" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.btpservice_cicd_abort_build.uut]
    }
  }
}

action "btpservice_cicd_abort_build" "uut" {
  config {
    job   = %q
    build = %d
  }
}
`, job, build)
}
