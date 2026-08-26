# This is a singleton resource — any non-empty string may be used as the import ID.
# terraform import btpservice_cicd_allowed_spaces.<resource_name> <any-string>

terraform import btpservice_cicd_allowed_spaces.this cicd_allowed_spaces

# terraform import using an import block

import {
  to = btpservice_cicd_allowed_spaces.this
  id = "cicd_allowed_spaces"
}
