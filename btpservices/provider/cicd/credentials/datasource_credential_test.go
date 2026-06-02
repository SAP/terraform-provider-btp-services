// btpservices/provider/cicd/credentials/datasource_basic_auth_test.go

package cicdcredentials_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/cicd/utils"
	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/tfutils"
)

func TestDatasourceCicdCredentialBasicAuth(t *testing.T) {
	t.Parallel()

	t.Run("read", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/datasource_credential_read")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
data "btpservice_cicd_credential" "uut" {
  name = "tf-ds-test-basic-auth"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_credential.uut", "id"),
						resource.TestCheckResourceAttr("data.btpservice_cicd_credential.uut", "name", "tf-ds-test-basic-auth"),
						resource.TestCheckResourceAttr("data.btpservice_cicd_credential.uut", "description", "Used by datasource test"),
					),
				},
			},
		})
	})

	t.Run("not_found", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/datasource_credential_not_found")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
data "btpservice_cicd_credential" "uut" {
  name = "this-credential-does-not-exist"
}
`,
					ExpectError: regexp.MustCompile(`Credential Not Found`),
				},
			},
		})
	})
}
