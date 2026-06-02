// btpservices/provider/cicd/jobs/list_resource_job_test.go

package cicdjobs_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/list"
	res "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/querycheck/queryfilter"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"

	cicdjobs "github.com/SAP/terraform-provider-btp-services/btpservices/provider/cicd/jobs"
	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/cicd/utils"
	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/tfutils"
)

func TestListResourceCicdJob(t *testing.T) {
	t.Parallel()

	t.Run("list all jobs", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/list_resource_job")
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
list "btpservice_cicd_job" "test" {
  provider = "btpservice"
}
`,
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLengthAtLeast("btpservice_cicd_job.test", 2),
					},
				},
				{
					Query: true,
					Config: utils.HCLProviderBlock(creds) + `
list "btpservice_cicd_job" "test" {
  provider         = "btpservice"
  include_resource = true
}
`,
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLengthAtLeast("btpservice_cicd_job.test", 2),
						querycheck.ExpectResourceKnownValues(
							"btpservice_cicd_job.test",
							queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
								"id": knownvalue.StringExact("ca9101f6-49b7-4e91-b632-e4c222fd9e4d"),
							}),
							[]querycheck.KnownValueCheck{
								{
									Path:       tfjsonpath.New("name"),
									KnownValue: knownvalue.NotNull(),
								},
								{
									Path:       tfjsonpath.New("pipeline"),
									KnownValue: knownvalue.NotNull(),
								},
								{
									Path:       tfjsonpath.New("pipeline_parameters"),
									KnownValue: knownvalue.NotNull(),
								},
							},
						),
					},
				},
			},
		})
	})

	t.Run("list filtered by pipeline", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/list_resource_job_pipeline_filter")
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
list "btpservice_cicd_job" "test" {
  provider         = "btpservice"
  include_resource = true
  config {
    pipeline = "cf-env"
  }
}
`,
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLengthAtLeast("btpservice_cicd_job.test", 2),
						querycheck.ExpectResourceKnownValues(
							"btpservice_cicd_job.test",
							queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
								"id": knownvalue.StringExact("ca9101f6-49b7-4e91-b632-e4c222fd9e4d"),
							}),
							[]querycheck.KnownValueCheck{
								{
									Path:       tfjsonpath.New("pipeline"),
									KnownValue: knownvalue.StringExact("cf-env"),
								},
							},
						),
					},
				},
			},
		})
	})

	t.Run("list filtered by repository_id", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/list_resource_job_repository_filter")
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
list "btpservice_cicd_job" "test" {
  provider         = "btpservice"
  include_resource = true
  config {
    repository_id = "4126058b-997f-45b3-9379-4a948b96949f"
  }
}
`,
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLengthAtLeast("btpservice_cicd_job.test", 2),
						querycheck.ExpectResourceKnownValues(
							"btpservice_cicd_job.test",
							queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
								"id": knownvalue.StringExact("ca9101f6-49b7-4e91-b632-e4c222fd9e4d"),
							}),
							[]querycheck.KnownValueCheck{
								{
									Path:       tfjsonpath.New("repository_id"),
									KnownValue: knownvalue.StringExact("4126058b-997f-45b3-9379-4a948b96949f"),
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

		r := cicdjobs.NewJobListResource().(list.ListResourceWithConfigure)
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
