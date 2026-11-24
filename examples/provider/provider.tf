terraform {
  required_version = ">= 1.14.0"

  required_providers {
    btpservice = {
      source  = "sap/btp-services"
      version = "1.0.0"
    }
  }
}

provider "btpservice" {
  cicd {
    endpoint      = "https://cicd-service-url.cfapps.us10.hana.ondemand.com"
    token_url     = "https://your-subaccount.authentication.us10.hana.ondemand.com/oauth/token"
    client_id     = "sb-clone-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx!b12345|cicd-service!b6789"
    client_secret = "your-client-secret-value-here="
  }

  ctms {
    endpoint      = "https://transport-service-app-backend.ts.cfapps.landscape.hana.ondemand.com"
    token_url     = "https://your-subaccount.authentication.us10.hana.ondemand.com/oauth/token"
    client_id     = "your-ctms-client-id"
    client_secret = "your-ctms-client-secret"
  }
}
