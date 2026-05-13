resource "btpservice_cicd_credential_cert_based_auth_custom_idp" "example" {
  name          = "my-cert-idp-user"
  description   = "Certificate-based auth credential for a custom identity provider"
  email_address = "deploy-user@example.com"
  hostname      = "my-idp.accounts.ondemand.com"
  origin        = "my-idp_platform"
}
