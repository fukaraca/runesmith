output "kubectl_context" {
  value = "kubectl config use-context ${var.profile}"
}

output "nodes" {
  value = [
    "minikube (control-plane)",
    "minikube-m02 (energy=fire)",
    "minikube-m03 (energy=frost)",
    "minikube-m04 (energy=arcane)",
  ]
}
