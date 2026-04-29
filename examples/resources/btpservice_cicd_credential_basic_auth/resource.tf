resource "btpservice_cicd_credential_basic_auth" "example" {
  name        = "my-deploy-user-from-cicd"
  description = "CF deployment user1"
  username    = "deploy-user@example.com"
  password    = "passoword"
}