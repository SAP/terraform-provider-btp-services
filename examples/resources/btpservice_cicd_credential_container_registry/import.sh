# terraform import btpservice_cicd_credential_container_registry.<resource_name> <credential_name>

terraform import btpservice_cicd_credential_container_registry.example my-container-registry

# terraform import using id attribute in import block

import {
  to = btpservice_cicd_credential_container_registry.<resource_name>
  id = "<credential_name>"
}

# this resource supports import using identity attribute from Terraform version 1.12 or higher

import {
  to = btpservice_cicd_credential_container_registry.<resource_name>
  identity = {
    id = "<credential_name>"
  }
}
