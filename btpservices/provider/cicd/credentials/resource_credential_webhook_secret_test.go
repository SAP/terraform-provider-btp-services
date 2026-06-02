// btpservices/provider/cicd/credentials/resource_credential_webhook_secret_test.go

package cicdcredentials_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/cicd/utils"
	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/tfutils"
)

func TestResourceCicdCredentialWebhookSecret(t *testing.T) {
	t.Parallel()

	t.Run("happy path - webhook secret creds", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/resource_credential_webhook_secret")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
resource "btpservice_cicd_credential_webhook_secret" "test" {
  name        = "tf-test-webhook-secret"
  description = "Terraform acceptance test webhook secret credential"
  token       = "my-secret-token"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("btpservice_cicd_credential_webhook_secret.test", "id"),
						resource.TestCheckResourceAttr("btpservice_cicd_credential_webhook_secret.test", "name", "tf-test-webhook-secret"),
						resource.TestCheckResourceAttr("btpservice_cicd_credential_webhook_secret.test", "description", "Terraform acceptance test webhook secret credential"),
					),
				},
				{
					// Step 2: Update description and token
					Config: utils.HCLProviderBlock(creds) + `
resource "btpservice_cicd_credential_webhook_secret" "test" {
  name        = "tf-test-webhook-secret"
  description = "Updated description"
  token       = "updated-secret-token"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("btpservice_cicd_credential_webhook_secret.test", "description", "Updated description"),
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
resource "btpservice_cicd_credential_webhook_secret" "test" {
  token = "my-secret-token"
}
`,
					ExpectError: regexp.MustCompile(`The argument "name" is required`),
				},
			},
		})
	})

	t.Run("error - missing token", func(t *testing.T) {
		t.Parallel()
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(utils.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(utils.Redacted) + `
resource "btpservice_cicd_credential_webhook_secret" "test" {
  name = "tf-test-missing-token"
}
`,
					ExpectError: regexp.MustCompile(`The argument "token" is required`),
				},
			},
		})
	})
}
