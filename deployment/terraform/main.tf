provider "null" {}

# 1) Start the control-plane node
resource "null_resource" "minikube_start" {
  triggers = {
    profile = var.profile
    ver     = var.kubernetes_version
    drv     = var.driver
    cpus    = var.cpus_master
    mem     = var.mem_master_mb
  }

  provisioner "local-exec" {
    interpreter = ["/bin/bash", "-lc"]
    command = <<-EOC
      set -euo pipefail
      # If the profile exists, this is a no-op
      minikube status -p ${var.profile} >/dev/null 2>&1 || true

      # Start control-plane with its own resources
      minikube start -p ${var.profile} \
        --driver=${var.driver} \
        --cpus=${var.cpus_master} \
        --memory=${var.mem_master_mb}mb \
        ${var.kubernetes_version != "" ? "--kubernetes-version=" ~ var.kubernetes_version : ""} \
        --extra-config=kubeadm.ignore-preflight-errors=NumCPU,Mem \
        --extra-config=kubelet.authentication-token-webhook=true \
        --extra-config=kubelet.read-only-port=0

      # Point kubectl context to this profile
      minikube -p ${var.profile} update-context
    EOC
  }

  # Cleanup on destroy
  provisioner "local-exec" {
    when    = destroy
    command = "minikube delete -p ${var.profile} || true"
    interpreter = ["/bin/bash", "-lc"]
  }
}

# 2) Add three worker nodes with smaller resources
resource "null_resource" "workers" {
  depends_on = [null_resource.minikube_start]

  triggers = {
    profile   = var.profile
    cpus      = var.cpus_worker
    mem       = var.mem_worker_mb
    signature = "add-3-workers"
  }

  provisioner "local-exec" {
    interpreter = ["/bin/bash", "-lc"]
    command = <<-EOC
      set -euo pipefail

      # Add workers m02, m03, m04 if they don't exist
      for i in 1 2 3; do
        NAME="minikube-m0$((i+1))"
        if ! minikube -p ${var.profile} kubectl -- get node "$NAME" >/dev/null 2>&1; then
          minikube -p ${var.profile} node add --worker \
            --cpus=${var.cpus_worker} \
            --memory=${var.mem_worker_mb}mb
        fi
      done

      # Show nodes for sanity
      minikube -p ${var.profile} kubectl -- get nodes -o wide
    EOC
  }
}

# 3) Label + taint workers with energy roles
resource "null_resource" "label_taint_workers" {
  depends_on = [null_resource.workers]

  triggers = {
    profile = var.profile
    sig     = "energy-taints-v1"
  }

  provisioner "local-exec" {
    interpreter = ["/bin/bash", "-lc"]
    command = <<-EOC
      set -euo pipefail
      # Expected worker names for docker driver
      declare -a NODES=( "minikube-m02" "minikube-m03" "minikube-m04" )
      declare -a ENERGIES=( "fire" "frost" "arcane" )

      for idx in "${!NODES[@]}"; do
        node="${NODES[$idx]}"
        energy="${ENERGIES[$idx]}"

        # Label for selector convenience
        minikube -p ${var.profile} kubectl -- label node "$node" energy=${energy} --overwrite
        minikube -p ${var.profile} kubectl -- label node "$node" manawell.io/energy=${energy} --overwrite

        # Taint so only matching workloads schedule
        minikube -p ${var.profile} kubectl -- taint node "$node" energy=${energy}:NoSchedule --overwrite || true
      done

      echo "Node roles applied:"
      minikube -p ${var.profile} kubectl -- get nodes --show-labels
    EOC
  }
}

# 4) Optional: show kubelet config info (purely informational)
resource "null_resource" "cluster_info" {
  depends_on = [null_resource.label_taint_workers]
  provisioner "local-exec" {
    interpreter = ["/bin/bash", "-lc"]
    command = <<-EOC
      set -euo pipefail
      echo "Context is set to:"
      kubectl config current-context
      echo
      kubectl get nodes -o wide
    EOC
  }
}
