// Configure the Google Cloud provider
provider "google" {
  credentials = file(var.cred_file)
  project     = var.project_name
  region      = var.region_name
}

resource "google_compute_disk" "cnb-disk" {
  for_each = toset(var.vm_name)
  name     = "${each.value}-disk"
  type     = "pd-ssd"
  zone     = var.location
  image    = "ubuntu-1804-bionic-v20200807"
  size     = var.disk_size
}

resource "google_compute_instance" "k8snode" {
  for_each     = toset(var.vm_name)
  name         = each.value
  machine_type = var.vm_size[index(var.vm_name, each.value)]
  zone         = var.location

  boot_disk {
    source = google_compute_disk.cnb-disk[each.key].name
  }

  metadata = {
    ssh-keys = "${var.user_name}:${file(var.ssh_keys)}"
  }

  network_interface {
    # A default network is created for all GCP projects
    network = "default"

    access_config {
      // Include this section to give the VM an external ip address
    }
  }
}

output "instance_ip_addresses" {
  value = {
    for instance in google_compute_instance.k8snode :
    instance.name => [instance.network_interface.0.access_config.0.nat_ip, instance.network_interface.0.network_ip]
  }
}
