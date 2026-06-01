resource "btpservice_cicd_credential_container_registry" "example" {
  name        = "my-container-registry"
  description = "Container registry credentials for Docker Hub"
  content = jsonencode({
    auths = {
      "registry.example.com" = {
        auth = base64encode("username:password")
      }
    }
  })
}
