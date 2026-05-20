package cicdrepositories_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	cicdrepositories "github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/cicd/repositories"
	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/cicd/utils"
	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/tfutils"
)

func TestDatasourceCicdRepositoryWebhookConfig(t *testing.T) {
	t.Parallel()

	t.Run("read", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/datasource_repository_webhook_config_read")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
data "btpservice_cicd_repository_webhook_config" "uut" {
  repository = "tf-ds-test-repo"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.btpservice_cicd_repository_webhook_config.uut", "repository", "tf-ds-test-repo"),
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_repository_webhook_config.uut", "webhook_uri"),
					),
				},
			},
		})
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/datasource_repository_webhook_config_not_found")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
data "btpservice_cicd_repository_webhook_config" "uut" {
  repository = "this-repository-does-not-exist"
}
`,
					ExpectError: regexp.MustCompile(`Repository Not Found`),
				},
			},
		})
	})

	t.Run("error path - configure", func(t *testing.T) {
		t.Parallel()

		d := cicdrepositories.NewRepositoryWebhookConfigDataSource().(datasource.DataSourceWithConfigure)
		resp := &datasource.ConfigureResponse{}
		req := datasource.ConfigureRequest{ProviderData: struct{}{}}
		d.Configure(context.Background(), req, resp)
		if !resp.Diagnostics.HasError() {
			t.Error("expected error for invalid provider data type")
		}
	})
}
