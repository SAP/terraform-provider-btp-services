// btpservices/provider/cicd/credentials/resource_credential_cert_based_auth_custom_idp_test.go

package cicdcredentials_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/cicd/utils"
	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/tfutils"
)

func TestAccResourceCicdCredentialCertBasedAuthCustomIdP(t *testing.T) {
	t.Parallel()

	t.Run("happy path - cert based auth custom idp creds", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/resource_credential_cert_based_auth_custom_idp")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
resource "btpservice_cicd_credential_cert_based_auth_custom_idp" "test" {
  name          = "tf-test-cert-cidp"
  description   = "Terraform acceptance test credential"
  email_address = "test-user@example.com"
  hostname      = "my-idp.accounts.ondemand.com"
  origin        = "my-idp_platform"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("btpservice_cicd_credential_cert_based_auth_custom_idp.test", "id"),
						resource.TestCheckResourceAttr("btpservice_cicd_credential_cert_based_auth_custom_idp.test", "name", "tf-test-cert-cidp"),
						resource.TestCheckResourceAttr("btpservice_cicd_credential_cert_based_auth_custom_idp.test", "description", "Terraform acceptance test credential"),
						resource.TestCheckResourceAttr("btpservice_cicd_credential_cert_based_auth_custom_idp.test", "email_address", "test-user@example.com"),
						resource.TestCheckResourceAttr("btpservice_cicd_credential_cert_based_auth_custom_idp.test", "hostname", "my-idp.accounts.ondemand.com"),
						resource.TestCheckResourceAttr("btpservice_cicd_credential_cert_based_auth_custom_idp.test", "origin", "my-idp_platform"),
					),
				},
				{
					// Step 2: Update description
					Config: utils.HCLProviderBlock(creds) + `
resource "btpservice_cicd_credential_cert_based_auth_custom_idp" "test" {
  name          = "tf-test-cert-cidp"
  description   = "Updated description"
  email_address = "test-user@example.com"
  hostname      = "my-idp.accounts.ondemand.com"
  origin        = "my-idp_platform"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("btpservice_cicd_credential_cert_based_auth_custom_idp.test", "description", "Updated description"),
						resource.TestCheckResourceAttr("btpservice_cicd_credential_cert_based_auth_custom_idp.test", "email_address", "test-user@example.com"),
					),
				},
				{
					// Step 3: Import by ID — all fields are returned by the API, no ignore needed
					ResourceName:      "btpservice_cicd_credential_cert_based_auth_custom_idp.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})

	t.Run("error - missing email_address", func(t *testing.T) {
		t.Parallel()
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(utils.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(utils.Redacted) + `
resource "btpservice_cicd_credential_cert_based_auth_custom_idp" "test" {
  name     = "tf-test-missing-email"
  hostname = "my-idp.accounts.ondemand.com"
  origin   = "my-idp_platform"
}
`,
					ExpectError: regexp.MustCompile(`The argument "email_address" is required`),
				},
			},
		})
	})

	t.Run("error - missing hostname", func(t *testing.T) {
		t.Parallel()
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(utils.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(utils.Redacted) + `
resource "btpservice_cicd_credential_cert_based_auth_custom_idp" "test" {
  name          = "tf-test-missing-hostname"
  email_address = "test@example.com"
  origin        = "my-idp_platform"
}
`,
					ExpectError: regexp.MustCompile(`The argument "hostname" is required`),
				},
			},
		})
	})
}
