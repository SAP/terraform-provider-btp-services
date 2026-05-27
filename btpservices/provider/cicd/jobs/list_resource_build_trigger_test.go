// btpservices/provider/cicd/jobs/list_resource_build_trigger_test.go

package cicdjobs_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/list"
	res "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/querycheck/queryfilter"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"

	cicdjobs "github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/cicd/jobs"
	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/cicd/utils"
	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/tfutils"
)

func TestListResourceCicdBuildTrigger(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/list_resource_build_trigger")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			TerraformVersionChecks: []tfversion.TerraformVersionCheck{
				tfversion.SkipBelow(tfversion.Version1_14_0),
			},
			Steps: []resource.TestStep{
				{
					Query: true,
					Config: utils.HCLProviderBlock(creds) + `
list "btpservice_cicd_build_trigger" "test" {
  provider = "btpservice"
  config {
    job = "tf-test-job"
  }
}
`,
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLengthAtLeast("btpservice_cicd_build_trigger.test", 1),
					},
				},
				{
					Query: true,
					Config: utils.HCLProviderBlock(creds) + `
list "btpservice_cicd_build_trigger" "test" {
  provider         = "btpservice"
  include_resource = true
  config {
    job = "tf-test-job"
  }
}
`,
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLengthAtLeast("btpservice_cicd_build_trigger.test", 1),
						querycheck.ExpectResourceKnownValues(
							"btpservice_cicd_build_trigger.test",
							queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
								"job": knownvalue.StringExact("tf-test-job"),
								"id":  knownvalue.StringRegexp(regexp.MustCompile(`^[0-9a-f-]{36}$`)),
							}),
							[]querycheck.KnownValueCheck{
								{
									Path:       tfjsonpath.New("type"),
									KnownValue: knownvalue.StringExact("timer"),
								},
								{
									Path:       tfjsonpath.New("timer").AtMapKey("branch"),
									KnownValue: knownvalue.StringExact("main"),
								},
								{
									Path:       tfjsonpath.New("timer").AtMapKey("cron"),
									KnownValue: knownvalue.StringExact("0 9 * * 1-5"),
								},
							},
						),
					},
				},
			},
		})
	})

	t.Run("error path - configure", func(t *testing.T) {
		t.Parallel()

		r := cicdjobs.NewBuildTriggerListResource().(list.ListResourceWithConfigure)
		resp := &res.ConfigureResponse{}
		req := res.ConfigureRequest{
			ProviderData: struct{}{},
		}

		r.Configure(context.Background(), req, resp)

		if !resp.Diagnostics.HasError() {
			t.Error("Expected error for invalid provider data type")
		}
	})
}
