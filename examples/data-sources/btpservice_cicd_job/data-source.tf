data "btpservice_cicd_job" "by_name" {
  name = "my-cf-env-job"
}

data "btpservice_cicd_job" "by_id" {
  id = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}
