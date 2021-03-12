provider "azurerm" {
  # The "feature" block is required for AzureRM provider 2.x.
  # If you're using version 1.x, the "features" block is not allowed.
  version = "~>2.0"
  features {}
}

resource "azurerm_resource_group" "myterraformgroup" {
  name     = var.rg_name
  location = var.location

  tags = {
    environment = "Terraform CNB"
  }
}

# Create virtual network
resource "azurerm_virtual_network" "myterraformnetwork" {
  name                = "${var.rg_name}-Vnet"
  address_space       = ["10.0.0.0/16"]
  location            = var.location
  resource_group_name = azurerm_resource_group.myterraformgroup.name

  tags = {
    environment = "Terraform CNB"
  }
}

# Create subnet
resource "azurerm_subnet" "myterraformsubnet" {
  name                 = "mySubnet"
  resource_group_name  = azurerm_resource_group.myterraformgroup.name
  virtual_network_name = azurerm_virtual_network.myterraformnetwork.name
  address_prefixes     = ["10.0.1.0/24"]
}

# Create public IPs
resource "azurerm_public_ip" "myterraformpublicip" {
  for_each            = toset(var.vm_name)
  name                = "${each.value}-PublicIP"
  location            = var.location
  resource_group_name = azurerm_resource_group.myterraformgroup.name
  allocation_method   = "Dynamic"

  tags = {
    environment = "Terraform CNB"
  }
}

# Create Network Security Group and rule
resource "azurerm_network_security_group" "myterraformnsg" {
  name                = "myNetworkSecurityGroup"
  location            = var.location
  resource_group_name = azurerm_resource_group.myterraformgroup.name

  security_rule {
    name                       = "SSH"
    priority                   = 1001
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "22"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }

  security_rule {
    name                       = "Internal"
    priority                   = 1000
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "*"
    source_port_range          = "*"
    destination_port_range     = "*"
    source_address_prefix      = "10.0.2.0/24"
    destination_address_prefix = "*"
  }

  tags = {
    environment = "Terraform CNB"
  }
}

# Create network interface
resource "azurerm_network_interface" "myterraformnic" {
  for_each            = toset(var.vm_name)
  name                = "${each.value}-nic"
  location            = var.location
  resource_group_name = azurerm_resource_group.myterraformgroup.name

  ip_configuration {
    name                          = "myNicConfiguration"
    subnet_id                     = azurerm_subnet.myterraformsubnet.id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = azurerm_public_ip.myterraformpublicip[each.key].id
  }

  tags = {
    environment = "Terraform CNB"
  }
}

# Connect the security group to the network interface
resource "azurerm_network_interface_security_group_association" "myAssociation" {
  for_each                  = toset(var.vm_name)
  network_interface_id      = azurerm_network_interface.myterraformnic[each.key].id
  network_security_group_id = azurerm_network_security_group.myterraformnsg.id
}

# Create virtual machine
resource "azurerm_linux_virtual_machine" "myterraformvm" {
  for_each              = toset(var.vm_name)
  name                  = each.value
  location              = var.location
  resource_group_name   = azurerm_resource_group.myterraformgroup.name
  network_interface_ids = [azurerm_network_interface.myterraformnic[each.key].id]
  size                  = var.vm_size[index(var.vm_name, each.value)]

  os_disk {
    name                 = "${each.value}-OsDisk"
    caching              = "ReadWrite"
    storage_account_type = var.storage_type
    disk_size_gb         = var.disk_size
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "18.04-LTS"
    version   = "latest"
  }

  computer_name                   = each.value
  admin_username                  = var.username
  disable_password_authentication = true

  admin_ssh_key {
    username   = var.username
    public_key = file(var.vm_sshkeyfile)
  }

  tags = {
    environment = "Terraform CNB"
  }
}

output "instance_ip_addresses" {
  value = {
    for instance in azurerm_linux_virtual_machine.myterraformvm :
    instance.name => [instance.public_ip_address, instance.private_ip_address]
  }
}
