resource "btpservice_cicd_credential_cloud_connector" "example" {
  name        = "my-cloud-connector"
  description = "Cloud Connector for on-premise system access"
  location_id = "my-location-id"
}
