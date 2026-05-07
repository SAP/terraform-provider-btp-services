resource "btpservice_cicd_credential_webhook_secret" "example" {
  name        = "my-webhook-secret"
  description = "Webhook secret for GitHub integration"
  token       = "my-secret-token"
}
