// btpservices/provider/cicd/repositories/resource_repository_test.go

package cicdrepositories_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/cicd/utils"
	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/tfutils"
)

func TestResourceCicdRepository(t *testing.T) {
	t.Parallel()

	t.Run("happy path - create and update", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/resource_repository")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
resource "btpservice_cicd_repository" "test" {
  name      = "tf-test-repo"
  clone_url = "https://github.com/example/tf-test-repo"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("btpservice_cicd_repository.test", "id"),
						resource.TestCheckResourceAttr("btpservice_cicd_repository.test", "name", "tf-test-repo"),
						resource.TestCheckResourceAttr("btpservice_cicd_repository.test", "clone_url", "https://github.com/example/tf-test-repo"),
					),
				},
				{
					// Step 2: Update clone_url
					Config: utils.HCLProviderBlock(creds) + `
resource "btpservice_cicd_repository" "test" {
  name      = "tf-test-repo"
  clone_url = "https://github.com/example/tf-test-repo-updated"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("btpservice_cicd_repository.test", "clone_url", "https://github.com/example/tf-test-repo-updated"),
					),
				},
				{
					// Step 3: Import by ID
					ResourceName:      "btpservice_cicd_repository.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})

	t.Run("happy path - with event_receiver", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/resource_repository_with_event_receiver")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
resource "btpservice_cicd_repository" "test" {
  name      = "tf-test-repo-webhook"
  clone_url = "https://github.com/example/tf-test-repo-webhook"

  event_receiver = {
    active                      = true
    scm_type                    = "GITHUB"
    webhook_token_credential_id = "dd1e05f6-01b6-40b2-b665-8cca9b89dc39"
  }
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("btpservice_cicd_repository.test", "id"),
						resource.TestCheckResourceAttr("btpservice_cicd_repository.test", "event_receiver.active", "true"),
						resource.TestCheckResourceAttr("btpservice_cicd_repository.test", "event_receiver.scm_type", "GITHUB"),
						resource.TestCheckResourceAttrSet("btpservice_cicd_repository.test", "event_receiver.webhook_id"),
					),
				},
				{
					// Step 2: Update event_receiver — webhook_id must be preserved in the PUT body
					Config: utils.HCLProviderBlock(creds) + `
resource "btpservice_cicd_repository" "test" {
  name      = "tf-test-repo-webhook"
  clone_url = "https://github.com/example/tf-test-repo-webhook"

  event_receiver = {
    active                      = false
    scm_type                    = "GITHUB"
    webhook_token_credential_id = "dd1e05f6-01b6-40b2-b665-8cca9b89dc39"
  }
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("btpservice_cicd_repository.test", "event_receiver.active", "false"),
						resource.TestCheckResourceAttrSet("btpservice_cicd_repository.test", "event_receiver.webhook_id"),
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
resource "btpservice_cicd_repository" "test" {
  clone_url = "https://github.com/example/repo"
}
`,
					ExpectError: regexp.MustCompile(`The argument "name" is required`),
				},
			},
		})
	})

	t.Run("error - missing clone_url", func(t *testing.T) {
		t.Parallel()
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(utils.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(utils.Redacted) + `
resource "btpservice_cicd_repository" "test" {
  name = "tf-test-repo"
}
`,
					ExpectError: regexp.MustCompile(`The argument "clone_url" is required`),
				},
			},
		})
	})
}
