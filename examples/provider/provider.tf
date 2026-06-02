terraform {
  required_providers {
    btpservice = {
      source = "sap/btp-services"
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
}