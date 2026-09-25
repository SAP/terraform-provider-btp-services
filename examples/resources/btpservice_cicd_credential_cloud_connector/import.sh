# terraform import btpservice_cicd_credential_cloud_connector.<resource_name> <credential_name>

terraform import btpservice_cicd_credential_cloud_connector.example my-cloud-connector

# terraform import using id attribute in import block

import {
  to = btpservice_cicd_credential_cloud_connector.<resource_name>
  id = "<credential_name>"
}

# this resource supports import using identity attribute from Terraform version 1.12 or higher

import {
  to = btpservice_cicd_credential_cloud_connector.<resource_name>
  identity = {
    id = "<credential_name>"
  }
}
