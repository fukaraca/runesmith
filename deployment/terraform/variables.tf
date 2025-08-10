variable "profile" {
  description = "Minikube profile name"
  type        = string
  default     = "runesmith"
}

variable "kubernetes_version" {
  description = "Kubernetes version for minikube (empty = default)"
  type        = string
  default     = ""
}

# Control-plane sizing
variable "cpus_master" { default = 1 }
variable "mem_master_mb" { default = 1024 } # 1 GiB is safer for control-plane

# Worker sizing
variable "cpus_worker" { default = 1 }
variable "mem_worker_mb" { default = 512 }

# Driver: docker works great with Docker Desktop
variable "driver" { default = "docker" }
