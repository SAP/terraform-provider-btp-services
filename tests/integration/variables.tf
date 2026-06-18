variable "github_repo_url" {
  description = "Public GitHub repository URL used for the test repository resource."
  type        = string
  default     = "https://github.com/SAP-samples/cloud-cf-helloworld-nodejs"
}

variable "github_webhook_secret" {
  description = "Webhook secret token for GitHub event receiver."
  type        = string
  sensitive   = true
  default     = "integration-test-webhook-secret"
}

variable "deploy_username" {
  description = "Username for the basic-auth deploy credential."
  type        = string
  default     = "integration-test-user@example.com"
}

variable "deploy_password" {
  description = "Password for the basic-auth deploy credential."
  type        = string
  sensitive   = true
  default     = "integration-test-password"
}
