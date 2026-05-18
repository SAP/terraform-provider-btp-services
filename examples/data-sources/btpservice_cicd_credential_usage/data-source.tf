# List all usages (jobs and repositories) for a credential.
data "btpservice_cicd_credential_usage" "all" {
  credential = "my-deploy-user"
}

# List only job usages for a credential.
data "btpservice_cicd_credential_usage" "jobs_only" {
  credential = "my-deploy-user"
  usertype   = "job"
}
