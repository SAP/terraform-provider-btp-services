resource "btpservice_cicd_build_trigger" "nightly" {
  job  = btpservice_cicd_repository.app.name
  type = "timer"

  timer = {
    branch = "main"
    cron   = "0 2 * * *"
  }
}

# Weekday morning build trigger
resource "btpservice_cicd_build_trigger" "weekday_morning" {
  job  = "my-pipeline-job"
  type = "timer"

  timer = {
    branch = "develop"
    cron   = "0 9 * * 1-5"
  }
}
