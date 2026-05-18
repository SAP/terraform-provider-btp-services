// btpservices/provider/cicd/repositories/datasource_repository_test.go

package cicdrepositories_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/cicd/utils"
	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/tfutils"
)

func TestDatasourceCicdRepository(t *testing.T) {
	t.Parallel()

	t.Run("read by name", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/datasource_repository_read")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
data "btpservice_cicd_repository" "uut" {
  name = "tf-ds-test-repo"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_repository.uut", "id"),
						resource.TestCheckResourceAttr("data.btpservice_cicd_repository.uut", "name", "tf-ds-test-repo"),
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_repository.uut", "clone_url"),
					),
				},
			},
		})
	})

	t.Run("not_found", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/datasource_repository_not_found")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
data "btpservice_cicd_repository" "uut" {
  name = "this-repository-does-not-exist"
}
`,
					ExpectError: regexp.MustCompile(`Repository Not Found`),
				},
			},
		})
	})
}
