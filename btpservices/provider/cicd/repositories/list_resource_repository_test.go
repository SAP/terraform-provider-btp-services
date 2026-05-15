// btpservices/provider/cicd/repositories/list_resource_repository_test.go

package cicdrepositories_test

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

	cicdrepositories "github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/cicd/repositories"
	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/cicd/utils"
	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/tfutils"
)

func TestListResourceCicdRepository(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/list_resource_repository")
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
list "btpservice_cicd_repository" "test" {
  provider = "btpservice"
}
`,
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLengthAtLeast("btpservice_cicd_repository.test", 1),
					},
				},
				{
					Query: true,
					Config: utils.HCLProviderBlock(creds) + `
list "btpservice_cicd_repository" "test" {
  provider         = "btpservice"
  include_resource = true
}
`,
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLengthAtLeast("btpservice_cicd_repository.test", 1),
						querycheck.ExpectResourceKnownValues(
							"btpservice_cicd_repository.test",
							queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
								"id": knownvalue.StringExact("4126058b-997f-45b3-9379-4a948b96949f"),
							}),
							[]querycheck.KnownValueCheck{
								{
									Path:       tfjsonpath.New("name"),
									KnownValue: knownvalue.StringExact("tf-ds-test-repo"),
								},
								{
									Path:       tfjsonpath.New("clone_url"),
									KnownValue: knownvalue.NotNull(),
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

		r := cicdrepositories.NewRepositoryListResource().(list.ListResourceWithConfigure)
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
