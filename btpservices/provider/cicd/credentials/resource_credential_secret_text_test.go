// btpservices/provider/cicd/credentials/resource_credential_secret_text_test.go

package cicdcredentials_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/cicd/utils"
	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/tfutils"
)

func TestAccResourceCicdCredentialSecretText(t *testing.T) {
	t.Parallel()

	t.Run("happy path - secret text creds", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/resource_credential_secret_text")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
resource "btpservice_cicd_credential_secret_text" "test" {
  name        = "tf-test-secret-text"
  description = "Terraform acceptance test credential"
  text        = "redacted-secret-value"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("btpservice_cicd_credential_secret_text.test", "id"),
						resource.TestCheckResourceAttr("btpservice_cicd_credential_secret_text.test", "name", "tf-test-secret-text"),
						resource.TestCheckResourceAttr("btpservice_cicd_credential_secret_text.test", "description", "Terraform acceptance test credential"),
					),
				},
				{
					// Step 2: Update description
					Config: utils.HCLProviderBlock(creds) + `
resource "btpservice_cicd_credential_secret_text" "test" {
  name        = "tf-test-secret-text"
  description = "Updated description"
  text        = "redacted-secret-value"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("btpservice_cicd_credential_secret_text.test", "description", "Updated description"),
					),
				},
			},
		})
	})

	t.Run("error - missing name", func(t *testing.T) {
		t.Parallel()
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(utils.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(utils.Redacted) + `
resource "btpservice_cicd_credential_secret_text" "test" {
  text = "some-secret"
}
`,
					ExpectError: regexp.MustCompile(`The argument "name" is required`),
				},
			},
		})
	})

	t.Run("error - missing text", func(t *testing.T) {
		t.Parallel()
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(utils.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(utils.Redacted) + `
resource "btpservice_cicd_credential_secret_text" "test" {
  name = "tf-test-missing-value"
}
`,
					ExpectError: regexp.MustCompile(`The argument "text" is required`),
				},
			},
		})
	})
}
