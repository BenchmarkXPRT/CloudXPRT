variable "vm_name" {
  description = "Name of the VMs"
  type        = list(string)
  default     = ["cnbvm1", "cnbvm2"]
}

variable "vm_size" {
  description = "Size of the VMs"
  type        = list(string)
  default     = ["n1-standard-4", "n1-standard-4"]
}

variable "disk_size" {
  description = "VM disk size in GB"
  default     = 50
}

variable "location" {
  description = "Region to build into"
  default     = "us-west1-b"
}

variable "region_name" {
  default = "us-west1"
}

variable "project_name" {
  default = "cnb-gcp-XXXXX"
}

variable "user_name" {
  default = "gcpuser"
}

variable "cred_file" {
  default = "CNB-GCP-XXXXXX.json"
}

variable "ssh_keys" {
  default = "~/.ssh/id_rsa.pub"
}
