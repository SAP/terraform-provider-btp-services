// main.go

// Run "make generate" to regenerate provider documentation.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name btpservice --provider-dir .

package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	btpservicesprovider "github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider"
)

// version is set by goreleaser at build time via -ldflags.
var version = "dev"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/SAP/sap-btp-services",
		Debug:   debug,
	}

	if err := providerserver.Serve(context.Background(), btpservicesprovider.New(version), opts); err != nil {
		log.Fatal(err)
	}
}
