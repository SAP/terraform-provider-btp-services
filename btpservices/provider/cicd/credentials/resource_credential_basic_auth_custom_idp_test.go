// btpservices/provider/cicd/credentials/resource_credential_basic_auth_custom_idp_test.go

package cicdcredentials_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/cicd/utils"
	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/tfutils"
)

func TestResourceCicdCredentialBasicAuthCustomIdP(t *testing.T) {
	t.Parallel()

	t.Run("happy path - basic auth custom idp creds", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/resource_credential_basic_auth_custom_idp")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
resource "btpservice_cicd_credential_basic_auth_custom_idp" "test" {
  name        = "tf-test-basic-auth-cidp"
  description = "Terraform acceptance test credential"
  username    = "test-user"
  password    = "test-password"
  origin      = "custom-platform"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("btpservice_cicd_credential_basic_auth_custom_idp.test", "id"),
						resource.TestCheckResourceAttr("btpservice_cicd_credential_basic_auth_custom_idp.test", "name", "tf-test-basic-auth-cidp"),
						resource.TestCheckResourceAttr("btpservice_cicd_credential_basic_auth_custom_idp.test", "description", "Terraform acceptance test credential"),
						resource.TestCheckResourceAttr("btpservice_cicd_credential_basic_auth_custom_idp.test", "username", "test-user"),
						resource.TestCheckResourceAttr("btpservice_cicd_credential_basic_auth_custom_idp.test", "origin", "custom-platform"),
					),
				},
				{
					// Step 2: Update description and username
					Config: utils.HCLProviderBlock(creds) + `
resource "btpservice_cicd_credential_basic_auth_custom_idp" "test" {
  name        = "tf-test-basic-auth-cidp"
  description = "Updated description"
  username    = "updated-user"
  password    = "test-password"
  origin      = "custom-platform"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("btpservice_cicd_credential_basic_auth_custom_idp.test", "description", "Updated description"),
						resource.TestCheckResourceAttr("btpservice_cicd_credential_basic_auth_custom_idp.test", "username", "updated-user"),
					),
				},
				{
					// Step 3: Import by ID — password excluded because API never returns it
					ResourceName:            "btpservice_cicd_credential_basic_auth_custom_idp.test",
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"password"},
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
resource "btpservice_cicd_credential_basic_auth_custom_idp" "test" {
  username = "test-user"
  password = "test-password"
  origin   = "custom-platform"
}
`,
					ExpectError: regexp.MustCompile(`The argument "name" is required`),
				},
			},
		})
	})

	t.Run("error - missing origin", func(t *testing.T) {
		t.Parallel()
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(utils.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(utils.Redacted) + `
resource "btpservice_cicd_credential_basic_auth_custom_idp" "test" {
  name     = "tf-test-missing-origin"
  username = "test-user"
  password = "test-password"
}
`,
					ExpectError: regexp.MustCompile(`The argument "origin" is required`),
				},
			},
		})
	})
}
