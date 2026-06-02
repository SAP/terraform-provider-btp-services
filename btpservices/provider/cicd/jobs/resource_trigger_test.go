// btpservices/provider/cicd/jobs/resource_trigger_test.go

package cicdjobs_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	cicdjobs "github.com/SAP/terraform-provider-btp-services/btpservices/provider/cicd/jobs"
	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/cicd/utils"
	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/tfutils"
	"github.com/SAP/terraform-provider-btp-services/internal/shared"
)

func TestResourceCicdTrigger(t *testing.T) {
	t.Parallel()

	t.Run("happy path - create, update, and import", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/resource_trigger")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
resource "btpservice_cicd_trigger" "test" {
  job  = "tf-test-job"
  type = "timer"
  timer = {
    branch = "main"
    cron   = "0 9 * * 1-5"
  }
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("btpservice_cicd_trigger.test", "id"),
						resource.TestCheckResourceAttr("btpservice_cicd_trigger.test", "job", "tf-test-job"),
						resource.TestCheckResourceAttr("btpservice_cicd_trigger.test", "type", "timer"),
						resource.TestCheckResourceAttr("btpservice_cicd_trigger.test", "timer.branch", "main"),
						resource.TestCheckResourceAttr("btpservice_cicd_trigger.test", "timer.cron", "0 9 * * 1-5"),
					),
				},
				{
					// Step 2: Update cron schedule
					Config: utils.HCLProviderBlock(creds) + `
resource "btpservice_cicd_trigger" "test" {
  job  = "tf-test-job"
  type = "timer"
  timer = {
    branch = "main"
    cron   = "0 10 * * 1-5"
  }
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("btpservice_cicd_trigger.test", "timer.cron", "0 10 * * 1-5"),
					),
				},
				{
					// Step 3: Import via composite "job/trigger_id" key derived from state
					ResourceName:      "btpservice_cicd_trigger.test",
					ImportState:       true,
					ImportStateVerify: true,
					ImportStateIdFunc: func(s *terraform.State) (string, error) {
						rs := s.RootModule().Resources["btpservice_cicd_trigger.test"]
						if rs == nil {
							return "", fmt.Errorf("resource not found in state")
						}
						return rs.Primary.Attributes["job"] + "," + rs.Primary.Attributes["id"], nil
					},
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
resource "btpservice_cicd_trigger" "test" {
  type = "timer"
  timer = {
    cron = "0 9 * * 1-5"
  }
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
resource "btpservice_cicd_trigger" "test" {
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
resource "btpservice_cicd_trigger" "test" {
  job  = "tf-test-job"
  type = "TIMER"
}
`,
					ExpectError: regexp.MustCompile(`value must be one of`),
				},
			},
		})
	})

	t.Run("error - timer block required when type is timer", func(t *testing.T) {
		t.Parallel()
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(utils.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(utils.Redacted) + `
resource "btpservice_cicd_trigger" "test" {
  job  = "tf-test-job"
  type = "timer"
}
`,
					ExpectError: regexp.MustCompile(`timer block is required`),
				},
			},
		})
	})

	t.Run("error - nil cicd client", func(t *testing.T) {
		t.Parallel()
		r := cicdjobs.NewTriggerResource().(fwresource.ResourceWithConfigure)
		resp := &fwresource.ConfigureResponse{}
		r.Configure(context.Background(), fwresource.ConfigureRequest{ProviderData: &shared.ProviderClients{Cicd: nil}}, resp)
		if !resp.Diagnostics.HasError() {
			t.Error("expected error when Cicd client is nil")
		}
	})
}
