variable "aws_region" {
  default = "us-west-2"
}

# For us-east-2 Ubuntu 18 is ami-0dd9f0e7df0f0a138
variable "instance_ami" {
  default = "ami-003634241a8fcdec0"
}

variable "vm_name" {
  type    = list(string)
  default = ["cnb-node1", "cnb-node2"]
}

variable "instance_type" {
  type    = list(string)
  default = ["m5.xlarge", "m5.xlarge"]
}

variable "vm_key" {
  default = "cnb_aws_key"
}

# For ubuntu VM, user name is always ubuntu
variable "vm_user" {
  default = "ubuntu"
}

variable "vm_volume" {
  default = "50"
}

variable "vpc_name" {
  default = "TerraformVPC"
}

variable "vpc_cidr" {
  default = "10.0.0.0/16"
}

variable "subnets_cidr" {
  type    = list(string)
  default = ["10.0.0.0/24"]
}

variable "ingress_cidr" {
  default = "0.0.0.0/0"
}

variable "aws_zones" {
  type    = list(string)
  default = ["us-west-2a"]
}
