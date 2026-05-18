resource "btpservice_cicd_repository" "example" {
  name      = "my-app-repository"
  clone_url = "https://github.com/example/my-app"
}

# Private repository with clone credential
resource "btpservice_cicd_repository" "private" {
  name               = "my-private-repo"
  clone_url          = "https://github.com/example/my-private-app"
  clone_credential_id = btpservice_cicd_credential_basic_auth.deploy_user.id
}

# Repository with webhook event receiver for automated builds
resource "btpservice_cicd_repository" "with_webhook" {
  name      = "my-webhook-repo"
  clone_url = "https://github.com/example/my-app"

  event_receiver = {
    active                      = true
    scm_type                    = "GITHUB"
    webhook_token_credential_id = btpservice_cicd_credential_webhook_secret.webhook.id
  }
}
