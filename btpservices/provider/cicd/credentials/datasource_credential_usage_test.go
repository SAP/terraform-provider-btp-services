package cicdcredentials_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	cicdcredentials "github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/cicd/credentials"
	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/cicd/utils"
	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/tfutils"
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
  credential = "tf-test-basic-auth-cidp"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.btpservice_cicd_credential_usage.uut", "credential", "tf-test-basic-auth-cidp"),
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_credential_usage.uut", "usages.#"),
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
  credential = "tf-test-basic-auth-cidp"
  usertype   = "job"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.btpservice_cicd_credential_usage.uut", "credential", "tf-test-basic-auth-cidp"),
						resource.TestCheckResourceAttr("data.btpservice_cicd_credential_usage.uut", "usertype", "job"),
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_credential_usage.uut", "usages.#"),
					),
				},
			},
		})
	})

	t.Run("error path - configure", func(t *testing.T) {
		t.Parallel()

		d := cicdcredentials.NewCredentialUsageDataSource().(datasource.DataSourceWithConfigure)
		resp := &datasource.ConfigureResponse{}
		req := datasource.ConfigureRequest{ProviderData: struct{}{}}
		d.Configure(context.Background(), req, resp)
		if !resp.Diagnostics.HasError() {
			t.Error("expected error for invalid provider data type")
		}
	})
}
