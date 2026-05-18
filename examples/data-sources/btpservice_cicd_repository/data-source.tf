data "btpservice_cicd_repository" "by_name" {
  name = "my-app-repository"
}

data "btpservice_cicd_repository" "by_id" {
  id = "pb091fd5-845b-4146-9bfc-d8cb74be04f8"
}
