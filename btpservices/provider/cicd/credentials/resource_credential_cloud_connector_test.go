// btpservices/provider/cicd/credentials/resource_credential_cloud_connector_test.go

package cicdcredentials_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/cicd/utils"
	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/tfutils"
)

func TestResourceCicdCredentialCloudConnector(t *testing.T) {
	t.Parallel()

	t.Run("happy path - cloud connector creds", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/resource_credential_cloud_connector")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
resource "btpservice_cicd_credential_cloud_connector" "test" {
  name        = "tf-test-cloud-connector"
  description = "Terraform acceptance test cloud connector credential"
  location_id = "my-location-id"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("btpservice_cicd_credential_cloud_connector.test", "id"),
						resource.TestCheckResourceAttr("btpservice_cicd_credential_cloud_connector.test", "name", "tf-test-cloud-connector"),
						resource.TestCheckResourceAttr("btpservice_cicd_credential_cloud_connector.test", "description", "Terraform acceptance test cloud connector credential"),
						resource.TestCheckResourceAttr("btpservice_cicd_credential_cloud_connector.test", "location_id", "my-location-id"),
					),
				},
				{
					// Step 2: Update description and location_id
					Config: utils.HCLProviderBlock(creds) + `
resource "btpservice_cicd_credential_cloud_connector" "test" {
  name        = "tf-test-cloud-connector"
  description = "Updated description"
  location_id = "updated-location-id"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("btpservice_cicd_credential_cloud_connector.test", "description", "Updated description"),
						resource.TestCheckResourceAttr("btpservice_cicd_credential_cloud_connector.test", "location_id", "updated-location-id"),
					),
				},
				{
					// Step 3: Import by ID
					ResourceName:      "btpservice_cicd_credential_cloud_connector.test",
					ImportState:       true,
					ImportStateVerify: true,
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
resource "btpservice_cicd_credential_cloud_connector" "test" {
  location_id = "my-location-id"
}
`,
					ExpectError: regexp.MustCompile(`The argument "name" is required`),
				},
			},
		})
	})

	t.Run("error - missing location_id", func(t *testing.T) {
		t.Parallel()
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(utils.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(utils.Redacted) + `
resource "btpservice_cicd_credential_cloud_connector" "test" {
  name = "tf-test-missing-location"
}
`,
					ExpectError: regexp.MustCompile(`The argument "location_id" is required`),
				},
			},
		})
	})
}
