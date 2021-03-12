variable "rg_name" {
  description = "Name of resource group"
  default     = "CNBTerraform-rg"
}

variable "location" {
  description = "Region to build into"
  default     = "eastus"
}

variable "vm_name" {
  description = "Name of the VMs"
  type        = list(string)
  default     = ["cnbvm1", "cnbvm2"]
}

variable "vm_size" {
  description = "Size of the VMs"
  type        = list(string)
  default     = ["Standard_D16s_v3", "Standard_D16s_v3"]
}

variable "storage_type" {
  description = "VM storage type"
  # D16a_v4 use Standard_LRS, D16as_v4 use Premium_LRS
  default = "Premium_LRS"
}

variable "disk_size" {
  description = "OS disk size in GB"
  default     = 50
}

variable "vm_sshkeyfile" {
  description = "SSH key file location"
  default     = "~/.ssh/id_rsa.pub"
}

variable "username" {
  description = "user name to create VM"
  default     = "azureuser"
}
