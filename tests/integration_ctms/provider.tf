terraform {
  required_providers {
    btpservice = {
      source = "SAP/btp-services"
    }
  }
}

# Credentials are read from environment variables:
#   BTP_CTMS_ENDPOINT
#   BTP_CTMS_TOKEN_URL
#   BTP_CTMS_CLIENT_ID
#   BTP_CTMS_CLIENT_SECRET
provider "btpservice" {
  ctms {}
}
