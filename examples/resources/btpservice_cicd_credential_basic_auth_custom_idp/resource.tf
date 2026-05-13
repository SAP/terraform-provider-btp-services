resource "btpservice_cicd_credential_basic_auth_custom_idp" "example" {
  name        = "my-custom-idp-user"
  description = "Basic auth credential for a custom identity provider"
  username    = "deploy-user@example.com"
  password    = "my-secret-password"
  origin      = "custom-platform"
}
