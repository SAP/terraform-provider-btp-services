// btpservices/provider/cicd/credentials/resource_credential_service_key_test.go

package cicdcredentials_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/cicd/utils"
	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/tfutils"
)

func TestAccResourceCicdCredentialServiceKey(t *testing.T) {
	t.Parallel()

	t.Run("happy path - service key creds", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/resource_credential_service_key")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
resource "btpservice_cicd_credential_service_key" "test" {
  name        = "tf-test-service-key"
  description = "Terraform acceptance test credential"
  key         = "{\"uri\":\"https://transport-service-app-backend.ts.cfapps.sap.hana.ondemand.com\",\"ua\":{\"uaadomain\":\"authentication.sap.hana.ondemand.com\",\"tenantode\":\"dedicated\"}}"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("btpservice_cicd_credential_service_key.test", "id"),
						resource.TestCheckResourceAttr("btpservice_cicd_credential_service_key.test", "name", "tf-test-service-key"),
						resource.TestCheckResourceAttr("btpservice_cicd_credential_service_key.test", "description", "Terraform acceptance test credential"),
					),
				},
				{
					// Step 2: Update description
					Config: utils.HCLProviderBlock(creds) + `
				resource "btpservice_cicd_credential_service_key" "test" {
				  name        = "tf-test-service-key"
				  description = "Updated description"
				  key         = "{\"uri\":\"https://transport-service-app-backend.ts.cfapps.sap.hana.ondemand.com\",\"ua\":{\"uaadomain\":\"authentication.sap.hana.ondemand.com\",\"tenantode\":\"dedicated\"}}"
				}
				`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("btpservice_cicd_credential_service_key.test", "description", "Updated description"),
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
resource "btpservice_cicd_credential_service_key" "test" {
  key = "some-service-key-json"
}
`,
					ExpectError: regexp.MustCompile(`The argument "name" is required`),
				},
			},
		})
	})

	t.Run("error - missing key", func(t *testing.T) {
		t.Parallel()
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(utils.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(utils.Redacted) + `
resource "btpservice_cicd_credential_service_key" "test" {
  name = "tf-test-missing-content"
}
`,
					ExpectError: regexp.MustCompile(`The argument "key" is required`),
				},
			},
		})
	})
}
