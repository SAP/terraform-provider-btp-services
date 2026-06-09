# Get the latest running (or most recently triggered) build.
data "btpservice_cicd_builds" "latest" {
  job    = "my-pipeline-job"
  filter = "latest"
}

# Get the most recently completed build.
data "btpservice_cicd_builds" "last_finished" {
  job    = "my-pipeline-job"
  filter = "latestFinished"
}