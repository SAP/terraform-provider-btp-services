terraform {
  required_providers {
    btpservice = {
      source = "SAP/btp-services"
    }
  }
}

# Credentials are read from environment variables:
#   BTP_CICD_ENDPOINT
#   BTP_CICD_TOKEN_URL
#   BTP_CICD_CLIENT_ID
#   BTP_CICD_CLIENT_SECRET
provider "btpservice" {
  cicd {}
}
