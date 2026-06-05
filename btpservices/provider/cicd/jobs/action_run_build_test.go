// btpservices/provider/cicd/jobs/action_run_build_test.go

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

func TestActionRunBuild(t *testing.T) {
	t.Parallel()

	t.Run("triggers build with minimal config (no guards)", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/action_run_build")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest: true,
			TerraformVersionChecks: []tfversion.TerraformVersionCheck{
				tfversion.SkipBelow(tfversion.Version1_14_0),
			},
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + hclActionRunBuildMinimal("tf-test-job"),
				},
			},
		})
	})

	t.Run("triggers build with etag, commit, and parameters", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/action_run_build_full")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest: true,
			TerraformVersionChecks: []tfversion.TerraformVersionCheck{
				tfversion.SkipBelow(tfversion.Version1_14_0),
			},
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + hclActionRunBuildFull("tf-test-job"),
				},
			},
		})
	})

	t.Run("error on 409 conflict (stale etag)", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/action_run_build_stale_etag")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest: true,
			TerraformVersionChecks: []tfversion.TerraformVersionCheck{
				tfversion.SkipBelow(tfversion.Version1_14_0),
			},
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config:      utils.HCLProviderBlock(creds) + hclActionRunBuildWithETag("tf-test-job", `W/"stale"`),
					ExpectError: regexp.MustCompile(`Build Not Triggered`),
				},
			},
		})
	})

	t.Run("error on invalid visibility value", func(t *testing.T) {
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
action "btpservice_cicd_run_build" "uut" {
  config {
    job = "tf-test-job"
    parameters = [
      {
        name       = "p"
        value      = "v"
        visibility = "INVALID"
      }
    ]
  }
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
			IsUnitTest: true,
			TerraformVersionChecks: []tfversion.TerraformVersionCheck{
				tfversion.SkipBelow(tfversion.Version1_14_0),
			},
			ProtoV6ProviderFactories: utils.GetTestProviders(utils.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(utils.Redacted) + `
action "btpservice_cicd_run_build" "uut" {
  config {}
}
`,
					ExpectError: regexp.MustCompile(`(?i)Missing required argument`),
				},
			},
		})
	})
}

func hclActionRunBuildMinimal(job string) string {
	return fmt.Sprintf(`
resource "terraform_data" "trigger" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.btpservice_cicd_run_build.uut]
    }
  }
}

action "btpservice_cicd_run_build" "uut" {
  config {
    commit_to_be_built = "main"
    job = %q
  }
}
`, job)
}

func hclActionRunBuildFull(job string) string {
	return fmt.Sprintf(`
resource "terraform_data" "trigger" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.btpservice_cicd_run_build.uut]
    }
  }
}

action "btpservice_cicd_run_build" "uut" {
  config {
    job                = %q
    commit_to_be_built = "main"
    parameters = [
      {
        name       = "addon.yml"
        value      = "enabled: true"
        visibility = "RESTRICTED"
      }
    ]
  }
}
`, job)
}

func hclActionRunBuildWithETag(job, etag string) string {
	return fmt.Sprintf(`
resource "terraform_data" "trigger" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.btpservice_cicd_run_build.uut]
    }
  }
}

action "btpservice_cicd_run_build" "uut" {
  config {
    job      = %q
    job_etag = %q
	commit_to_be_built = "main"
  }
}
`, job, etag)
}
