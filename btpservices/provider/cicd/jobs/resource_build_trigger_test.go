// btpservices/provider/cicd/jobs/resource_build_trigger_test.go

package cicdjobs_test

import (
	"context"
	"regexp"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	cicdjobs "github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/cicd/jobs"
	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/cicd/utils"
	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/tfutils"
	"github.com/SAP/terraform-provider-sap-btp-services/internal/shared"
)

func TestResourceCicdBuildTrigger(t *testing.T) {
	t.Parallel()

	t.Run("happy path - timer trigger", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/resource_build_trigger")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
resource "btpservice_cicd_build_trigger" "test" {
  job  = "tf-test-job"
  type = "timer"
  timer = {
    branch = "main"
    cron   = "0 9 * * 1-5"
  }
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("btpservice_cicd_build_trigger.test", "id"),
						resource.TestCheckResourceAttr("btpservice_cicd_build_trigger.test", "job", "tf-test-job"),
						resource.TestCheckResourceAttr("btpservice_cicd_build_trigger.test", "type", "timer"),
						resource.TestCheckResourceAttr("btpservice_cicd_build_trigger.test", "timer.branch", "main"),
						resource.TestCheckResourceAttr("btpservice_cicd_build_trigger.test", "timer.cron", "0 9 * * 1-5"),
					),
				},
			},
		})
	})

	t.Run("error - missing job", func(t *testing.T) {
		t.Parallel()
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(utils.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(utils.Redacted) + `
resource "btpservice_cicd_build_trigger" "test" {
  type = "TIMER"
}
`,
					ExpectError: regexp.MustCompile(`The argument "job" is required`),
				},
			},
		})
	})

	t.Run("error - missing type", func(t *testing.T) {
		t.Parallel()
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(utils.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(utils.Redacted) + `
resource "btpservice_cicd_build_trigger" "test" {
  job = "tf-test-job"
}
`,
					ExpectError: regexp.MustCompile(`The argument "type" is required`),
				},
			},
		})
	})

	t.Run("error - invalid type", func(t *testing.T) {
		t.Parallel()
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(utils.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(utils.Redacted) + `
resource "btpservice_cicd_build_trigger" "test" {
  job  = "tf-test-job"
  type = "TIMER"
}
`,
					ExpectError: regexp.MustCompile(`value must be one of`),
				},
			},
		})
	})

	t.Run("error - nil cicd client", func(t *testing.T) {
		t.Parallel()
		r := cicdjobs.NewBuildTriggerResource().(fwresource.ResourceWithConfigure)
		resp := &fwresource.ConfigureResponse{}
		r.Configure(context.Background(), fwresource.ConfigureRequest{ProviderData: &shared.ProviderClients{Cicd: nil}}, resp)
		if !resp.Diagnostics.HasError() {
			t.Error("expected error when Cicd client is nil")
		}
	})
}
