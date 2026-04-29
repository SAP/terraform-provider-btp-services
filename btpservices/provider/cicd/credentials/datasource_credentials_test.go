// btpservices/provider/services/cicd/credentials/datasource_credentials_test.go

package cicdcredentials_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/cicd/cicdtest"
	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/testutil"
)

func TestAccDatasourceCicdCredentials(t *testing.T) {
	t.Parallel()

	t.Run("list", func(t *testing.T) {
		t.Parallel()

		rec, creds := cicdtest.SetupVCR(t, "fixtures/datasource_credentials_list")
		defer testutil.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: cicdtest.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: cicdtest.HCLProviderBlock(creds) + `
data "btpservice_cicd_credentials" "uut" {}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_credentials.uut", "id"),
						resource.TestCheckResourceAttrSet("data.btpservice_cicd_credentials.uut", "values.#"),
					),
				},
			},
		})
	})
}
