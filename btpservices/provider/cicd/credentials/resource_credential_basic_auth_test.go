// btpservices/provider/services/cicd/credentials/resource_basic_auth_test.go

package cicdcredentials_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/cicd/cicdtest"
	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/testutil"
)

func TestAccResourceCicdCredentialBasicAuth(t *testing.T) {
	t.Parallel()

	t.Run("happy path - basic creds", func(t *testing.T) {
		t.Parallel()

		rec, creds := cicdtest.SetupVCR(t, "fixtures/resource_credential_basic_auth")
		defer testutil.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: cicdtest.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: cicdtest.HCLProviderBlock(creds) + `
resource "btpservice_cicd_credential_basic_auth" "test" {
  name        = "tf-test-basic-auth"
  description = "Terraform acceptance test credential"
  username    = "test-user"
  password    = "test-password"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("btpservice_cicd_credential_basic_auth.test", "id"),
						resource.TestCheckResourceAttr("btpservice_cicd_credential_basic_auth.test", "name", "tf-test-basic-auth"),
						resource.TestCheckResourceAttr("btpservice_cicd_credential_basic_auth.test", "description", "Terraform acceptance test credential"),
						resource.TestCheckResourceAttr("btpservice_cicd_credential_basic_auth.test", "username", "test-user"),
					),
				},
				{
					// Step 2: Update description and username
					Config: cicdtest.HCLProviderBlock(creds) + `
resource "btpservice_cicd_credential_basic_auth" "test" {
  name        = "tf-test-basic-auth"
  description = "Updated description"
  username    = "updated-user"
  password    = "test-password"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("btpservice_cicd_credential_basic_auth.test", "description", "Updated description"),
						resource.TestCheckResourceAttr("btpservice_cicd_credential_basic_auth.test", "username", "updated-user"),
					),
				},
				{
					// Step 3: Import by ID — password excluded because API never returns it
					ResourceName:            "btpservice_cicd_credential_basic_auth.test",
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
			ProtoV6ProviderFactories: cicdtest.GetTestProviders(cicdtest.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: cicdtest.HCLProviderBlock(cicdtest.Redacted) + `
resource "btpservice_cicd_credential_basic_auth" "test" {
  username = "test-user"
  password = "test-password"
}
`,
					ExpectError: regexp.MustCompile(`The argument "name" is required`),
				},
			},
		})
	})

	t.Run("error - missing username", func(t *testing.T) {
		t.Parallel()
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: cicdtest.GetTestProviders(cicdtest.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: cicdtest.HCLProviderBlock(cicdtest.Redacted) + `
resource "btpservice_cicd_credential_basic_auth" "test" {
  name     = "tf-test-missing-user"
  password = "test-password"
}
`,
					ExpectError: regexp.MustCompile(`The argument "username" is required`),
				},
			},
		})
	})

	t.Run("error - missing password", func(t *testing.T) {
		t.Parallel()
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: cicdtest.GetTestProviders(cicdtest.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: cicdtest.HCLProviderBlock(cicdtest.Redacted) + `
resource "btpservice_cicd_credential_basic_auth" "test" {
  name     = "tf-test-missing-pass"
  username = "test-user"
}
`,
					ExpectError: regexp.MustCompile(`The argument "password" is required`),
				},
			},
		})
	})

}
