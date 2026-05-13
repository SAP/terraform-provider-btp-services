resource "btpservice_cicd_credential_secret_text" "example" {
  name        = "my-api-token"
  description = "API token for accessing an external service"
  text        = "my-secret-api-token"
}
