resource "btpservice_cicd_credential_kubernetes_config" "example" {
  name        = "my-kubernetes-config"
  description = "Kubeconfig for production cluster"
  content     = file("~/.kube/config")
}
