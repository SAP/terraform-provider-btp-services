resource "btpservice_cicd_credential_service_key" "example" {
  name        = "my-service-key"
  description = "Service key for a BTP service instance"
  key         = "{\"uri\":\"https://my-service.cfapps.sap.hana.ondemand.com\",\"uaa\":{\"clientid\":\"my-client\",\"clientsecret\":\"my-secret\",\"url\":\"https://my-subaccount.authentication.sap.hana.ondemand.com\"}}"
}
