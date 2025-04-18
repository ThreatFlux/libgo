#!/bin/bash
set -e

echo "[INFO] Setting up libvirt environment for integration testing..."

# Ensure we have access to libvirt directories
echo "[INFO] Setting up libvirt directories..."
if [ "$(id -u)" -ne 0 ]; then
    echo "[ERROR] This script must be run as root"
    exit 1
fi

# Add current user to libvirt group
usermod -a -G libvirt vtriple
usermod -a -G kvm vtriple

# Ensure /var/lib/libvirt/images exists and has correct permissions
mkdir -p /var/lib/libvirt/images
chown root:libvirt /var/lib/libvirt/images
chmod 775 /var/lib/libvirt/images

# Check if default storage pool exists
echo "[INFO] Setting up libvirt storage pool..."
if ! virsh pool-info default &>/dev/null; then
    echo "[INFO] Creating default storage pool..."
    virsh pool-define-as default dir --target /var/lib/libvirt/images
    virsh pool-build default
    virsh pool-start default
    virsh pool-autostart default
else
    echo "[INFO] Default storage pool already exists."
    # Update pool target if needed
    virsh pool-dumpxml default | grep -q "/var/lib/libvirt/images" || {
        echo "[INFO] Updating default pool target..."
        virsh pool-destroy default || true
        virsh pool-undefine default
        virsh pool-define-as default dir --target /var/lib/libvirt/images
        virsh pool-build default
        virsh pool-start default
        virsh pool-autostart default
    }
fi

# Check if default network exists
echo "[INFO] Checking libvirt network..."
if ! virsh net-info default &>/dev/null; then
    echo "[INFO] Default network doesn't exist, creating..."
    # Create a simple default network if it doesn't exist
    cat > /tmp/default-network.xml << EOF
<network>
  <name>default</name>
  <forward mode='nat'/>
  <bridge name='virbr0' stp='on' delay='0'/>
  <ip address='192.168.122.1' netmask='255.255.255.0'>
    <dhcp>
      <range start='192.168.122.2' end='192.168.122.254'/>
    </dhcp>
  </ip>
</network>
EOF
    virsh net-define /tmp/default-network.xml
    virsh net-start default
    virsh net-autostart default
    rm /tmp/default-network.xml
else
    echo "[INFO] Default network already exists."
    # Make sure it's running
    if ! virsh net-info default | grep -q "Active:.*yes"; then
        echo "[INFO] Starting default network..."
        virsh net-start default
    fi
fi

# Download Ubuntu 24.04 cloud image if it doesn't exist
UBUNTU_IMG="/var/lib/libvirt/images/ubuntu-2404-base.qcow2"
if [ ! -f "$UBUNTU_IMG" ]; then
    echo "[INFO] Downloading Ubuntu 24.04 cloud image..."
    wget -O /tmp/noble-server-cloudimg-amd64.img https://cloud-images.ubuntu.com/noble/current/noble-server-cloudimg-amd64.img
    echo "[INFO] Converting image to qcow2 format..."
    qemu-img convert -f qcow2 -O qcow2 /tmp/noble-server-cloudimg-amd64.img "$UBUNTU_IMG"
    rm /tmp/noble-server-cloudimg-amd64.img
    echo "[INFO] Ubuntu image prepared at $UBUNTU_IMG"
else
    echo "[INFO] Ubuntu image already exists at $UBUNTU_IMG"
fi

echo "[INFO] Libvirt environment setup complete!"
