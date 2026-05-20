// btpservices/provider/cicd/credentials/resource_credential_kubernetes_config_test.go

package cicdcredentials_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/cicd/utils"
	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/tfutils"
)

func TestResourceCicdCredentialKubernetesConfig(t *testing.T) {
	t.Parallel()

	t.Run("happy path - kubernetes config creds", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/resource_credential_kubernetes_config")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
resource "btpservice_cicd_credential_kubernetes_config" "test" {
  name        = "tf-test-kubernetes-config"
  description = "Terraform acceptance test kubernetes config credential"
  content     = "apiVersion: v1\nkind: Config\nclusters: []\ncontexts: []\ncurrent-context: \"\"\nusers: []"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("btpservice_cicd_credential_kubernetes_config.test", "id"),
						resource.TestCheckResourceAttr("btpservice_cicd_credential_kubernetes_config.test", "name", "tf-test-kubernetes-config"),
						resource.TestCheckResourceAttr("btpservice_cicd_credential_kubernetes_config.test", "description", "Terraform acceptance test kubernetes config credential"),
					),
				},
				{
					// Step 2: Update description and content
					Config: utils.HCLProviderBlock(creds) + `
resource "btpservice_cicd_credential_kubernetes_config" "test" {
  name        = "tf-test-kubernetes-config"
  description = "Updated description"
  content     = "apiVersion: v1\nkind: Config\nclusters: []\ncontexts: []\ncurrent-context: \"\"\nusers: []"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("btpservice_cicd_credential_kubernetes_config.test", "description", "Updated description"),
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
resource "btpservice_cicd_credential_kubernetes_config" "test" {
  content = "apiVersion: v1"
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
resource "btpservice_cicd_credential_kubernetes_config" "test" {
  name = "tf-test-missing-content"
}
`,
					ExpectError: regexp.MustCompile(`The argument "content" is required`),
				},
			},
		})
	})
}
