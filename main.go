//go:generate go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest
//go:generate tfplugindocs generate --rendered-provider-name "SAP BTP Services" --provider-name "btpservice"

package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	btpservicesprovider "github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	err := providerserver.Serve(context.Background(), btpservicesprovider.New(), providerserver.ServeOpts{
		Address:         "registry.terraform.io/sap/sap-btp-services",
		Debug:           debug,
		ProtocolVersion: 6,
	})

	if err != nil {
		log.Fatal(err)
	}
}
