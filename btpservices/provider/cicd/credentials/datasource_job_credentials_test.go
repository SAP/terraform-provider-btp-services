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

func TestDatasourceCicdJobCredentials(t *testing.T) {
	t.Parallel()

	t.Run("read", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/datasource_job_credentials_read")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
data "btpservice_cicd_job_credentials" "uut" {
  job = "tf-ds-test-job-credentials"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.btpservice_cicd_job_credentials.uut", "job", "tf-ds-test-job-credentials"),
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_job_credentials.uut", "credential_ids.#"),
					),
				},
			},
		})
	})

	t.Run("error path - configure", func(t *testing.T) {
		t.Parallel()

		d := cicdcredentials.NewJobCredentialsDataSource().(datasource.DataSourceWithConfigure)
		resp := &datasource.ConfigureResponse{}
		req := datasource.ConfigureRequest{ProviderData: struct{}{}}
		d.Configure(context.Background(), req, resp)
		if !resp.Diagnostics.HasError() {
			t.Error("expected error for invalid provider data type")
		}
	})
}
