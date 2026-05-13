# terraform import btpservice_cicd_credential_secret_text.<resource_name> <id>

terraform import btpservice_cicd_credential_secret_text.example dd005d8b-1fee-4e6b-b6ff-cb9a197b7fe0

import {
  to = btpservice_cicd_credential_secret_text.example
  id = "dd005d8b-1fee-4e6b-b6ff-cb9a197b7fe0"
}

import {
  to = btpservice_cicd_credential_secret_text.example
  identity = {
    id = "dd005d8b-1fee-4e6b-b6ff-cb9a197b7fe0"
  }
}
