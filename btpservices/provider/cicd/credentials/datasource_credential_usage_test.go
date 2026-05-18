package cicdcredentials_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	cicdcredentials "github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/cicd/credentials"
	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/cicd/utils"
	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/tfutils"
	"github.com/SAP/terraform-provider-sap-btp-services/internal/shared"
)

func TestDatasourceCicdCredentialUsage(t *testing.T) {
	t.Parallel()

	t.Run("read all usages", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/datasource_credential_usage_read")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
data "btpservice_cicd_credential_usage" "uut" {
  credential = "tf-ds-test-basic-auth"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.btpservice_cicd_credential_usage.uut", "credential", "tf-ds-test-basic-auth"),
						resource.TestCheckResourceAttr("data.btpservice_cicd_credential_usage.uut", "usages.#", "3"),
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_credential_usage.uut", "usages.0.id"),
						resource.TestCheckResourceAttr("data.btpservice_cicd_credential_usage.uut", "usages.0.name", "tf-test-job"),
						resource.TestCheckResourceAttr("data.btpservice_cicd_credential_usage.uut", "usages.0.type", "job"),
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_credential_usage.uut", "usages.1.id"),
						resource.TestCheckResourceAttr("data.btpservice_cicd_credential_usage.uut", "usages.1.name", "tf-test-repo"),
						resource.TestCheckResourceAttr("data.btpservice_cicd_credential_usage.uut", "usages.1.type", "repository"),
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_credential_usage.uut", "usages.2.id"),
						resource.TestCheckResourceAttr("data.btpservice_cicd_credential_usage.uut", "usages.2.name", "tf-ds-test-repo"),
						resource.TestCheckResourceAttr("data.btpservice_cicd_credential_usage.uut", "usages.2.type", "repository"),
					),
				},
			},
		})
	})

	t.Run("read job usages only", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/datasource_credential_usage_jobs")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
data "btpservice_cicd_credential_usage" "uut" {
  credential = "tf-ds-test-basic-auth"
  usertype   = "job"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.btpservice_cicd_credential_usage.uut", "credential", "tf-ds-test-basic-auth"),
						resource.TestCheckResourceAttr("data.btpservice_cicd_credential_usage.uut", "usertype", "job"),
						resource.TestCheckResourceAttr("data.btpservice_cicd_credential_usage.uut", "usages.#", "1"),
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_credential_usage.uut", "usages.0.id"),
						resource.TestCheckResourceAttr("data.btpservice_cicd_credential_usage.uut", "usages.0.name", "tf-test-job"),
						resource.TestCheckResourceAttr("data.btpservice_cicd_credential_usage.uut", "usages.0.type", "job"),
					),
				},
			},
		})
	})

	t.Run("error path - wrong provider data type", func(t *testing.T) {
		t.Parallel()

		d := cicdcredentials.NewCredentialUsageDataSource().(datasource.DataSourceWithConfigure)
		resp := &datasource.ConfigureResponse{}
		req := datasource.ConfigureRequest{ProviderData: struct{}{}}
		d.Configure(context.Background(), req, resp)
		if !resp.Diagnostics.HasError() {
			t.Error("expected error for invalid provider data type")
		}
	})

	t.Run("error path - nil cicd client", func(t *testing.T) {
		t.Parallel()

		d := cicdcredentials.NewCredentialUsageDataSource().(datasource.DataSourceWithConfigure)
		resp := &datasource.ConfigureResponse{}
		req := datasource.ConfigureRequest{ProviderData: &shared.ProviderClients{Cicd: nil}}
		d.Configure(context.Background(), req, resp)
		if !resp.Diagnostics.HasError() {
			t.Error("expected error when Cicd client is nil")
		}
	})
}
