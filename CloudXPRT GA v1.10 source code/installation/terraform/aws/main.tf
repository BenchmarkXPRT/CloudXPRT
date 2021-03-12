provider "aws" {
  region = var.aws_region
}

# VPC
resource "aws_vpc" "terra_vpc" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true

  tags = {
    Name = var.vpc_name
  }
}

# Internet Gateway
resource "aws_internet_gateway" "terra_igw" {
  vpc_id = aws_vpc.terra_vpc.id

  tags = {
    Name = "TerraformGateway"
  }
}

# Subnets : public
resource "aws_subnet" "public" {
  count             = length(var.subnets_cidr)
  vpc_id            = aws_vpc.terra_vpc.id
  cidr_block        = element(var.subnets_cidr, count.index)
  availability_zone = element(var.aws_zones, count.index)
  tags = {
    Name = "Subnet-${count.index + 1}"
  }
}

# Route table: attach Internet Gateway 
resource "aws_route_table" "public_rt" {
  vpc_id = aws_vpc.terra_vpc.id
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.terra_igw.id
  }
  tags = {
    Name = "publicRouteTable"
  }
}

# Route table association with public subnets
resource "aws_route_table_association" "a" {
  count          = length(var.subnets_cidr)
  subnet_id      = element(aws_subnet.public.*.id, count.index)
  route_table_id = aws_route_table.public_rt.id
}

resource "aws_security_group" "cluster_sg" {
  name        = "allow_traffic"
  description = "Allowed inbound traffic"
  vpc_id      = aws_vpc.terra_vpc.id

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = [var.ingress_cidr]
  }

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = var.subnets_cidr
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_key_pair" "mykey" {
  key_name   = var.vm_key
  public_key = file("~/.ssh/id_rsa.pub")
}

resource "aws_instance" "terraforminstance" {
  count                       = length(var.vm_name)
  ami                         = var.instance_ami
  instance_type               = var.instance_type[count.index]
  security_groups             = [aws_security_group.cluster_sg.id]
  subnet_id                   = element(aws_subnet.public.*.id, count.index)
  associate_public_ip_address = true

  key_name = aws_key_pair.mykey.key_name
  connection {
    type        = "ssh"
    user        = var.vm_user
    private_key = file("~/.ssh/id_rsa")
    host        = self.public_ip
  }

  root_block_device {
    volume_size = var.vm_volume
  }

  tags = {
    Name = var.vm_name[count.index]
  }
}

output "instance_ip_addresses" {
  value = {
    for instance in aws_instance.terraforminstance :
    instance.tags["Name"] => [instance.public_ip, instance.private_ip]
  }
}
