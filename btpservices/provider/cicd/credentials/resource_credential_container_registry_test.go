// btpservices/provider/cicd/credentials/resource_credential_container_registry_test.go

package cicdcredentials_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/cicd/utils"
	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/tfutils"
)

func TestResourceCicdCredentialContainerRegistry(t *testing.T) {
	t.Parallel()

	t.Run("happy path - container registry creds", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/resource_credential_container_registry")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
resource "btpservice_cicd_credential_container_registry" "test" {
  name        = "tf-test-container-registry"
  description = "Terraform acceptance test container registry credential"
  content     = "{\"auths\":{\"registry.example.com\":{\"auth\":\"dXNlcjpwYXNz\"}}}"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("btpservice_cicd_credential_container_registry.test", "id"),
						resource.TestCheckResourceAttr("btpservice_cicd_credential_container_registry.test", "name", "tf-test-container-registry"),
						resource.TestCheckResourceAttr("btpservice_cicd_credential_container_registry.test", "description", "Terraform acceptance test container registry credential"),
					),
				},
				{
					// Step 2: Update description and content
					Config: utils.HCLProviderBlock(creds) + `
resource "btpservice_cicd_credential_container_registry" "test" {
  name        = "tf-test-container-registry"
  description = "Updated description"
  content     = "{\"auths\":{\"registry.example.com\":{\"auth\":\"dXBkYXRlZA==\"}}}"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("btpservice_cicd_credential_container_registry.test", "description", "Updated description"),
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
resource "btpservice_cicd_credential_container_registry" "test" {
  content = "{}"
}
`,
					ExpectError: regexp.MustCompile(`The argument "name" is required`),
				},
			},
		})
	})

	t.Run("error - missing content", func(t *testing.T) {
		t.Parallel()
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(utils.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(utils.Redacted) + `
resource "btpservice_cicd_credential_container_registry" "test" {
  name = "tf-test-missing-content"
}
`,
					ExpectError: regexp.MustCompile(`The argument "content" is required`),
				},
			},
		})
	})
}
