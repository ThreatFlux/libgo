#!/bin/bash

# Windows VM Creation Script using LibGo API
# This script demonstrates creating a Windows VM from scratch

set -e

# Configuration
API_BASE_URL="${API_BASE_URL:-http://localhost:8080/api/v1}"
VM_NAME="${VM_NAME:-windows11-vm}"
VM_TEMPLATE="${VM_TEMPLATE:-windows-11}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Function to check API health
check_api_health() {
    log_info "Checking API health..."
    if curl -s -f "${API_BASE_URL%/api/v1}/health" > /dev/null; then
        log_info "API is healthy"
        return 0
    else
        log_error "API is not responding"
        return 1
    fi
}

# Function to create VM
create_windows_vm() {
    local vm_name=$1
    local template=$2
    
    log_info "Creating Windows VM: $vm_name"
    
    # VM Configuration
    local vm_config=$(cat <<EOF
{
    "name": "${vm_name}",
    "template": "${template}",
    "description": "Windows 11 VM created via API",
    "cpu": {
        "count": 4,
        "cores": 2,
        "threads": 1,
        "socket": 2,
        "model": "host-passthrough"
    },
    "memory": {
        "sizeBytes": 8589934592
    },
    "disk": {
        "sizeBytes": 107374182400,
        "format": "qcow2",
        "bus": "sata",
        "storagePool": "default"
    },
    "network": {
        "type": "network",
        "source": "default",
        "model": "virtio"
    },
    "cloudInit": {
        "enabled": false
    }
}
EOF
)
    
    # Create the VM
    response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$vm_config" \
        "${API_BASE_URL}/vms")
    
    # Check if creation was successful
    if echo "$response" | grep -q '"vm"'; then
        local uuid=$(echo "$response" | grep -o '"uuid":"[^"]*' | cut -d'"' -f4)
        log_info "VM created successfully!"
        log_info "UUID: $uuid"
        echo "$response" | jq . 2>/dev/null || echo "$response"
        return 0
    else
        log_error "Failed to create VM"
        echo "$response"
        return 1
    fi
}

# Function to start VM
start_vm() {
    local vm_name=$1
    
    log_info "Starting VM: $vm_name"
    
    response=$(curl -s -X PUT "${API_BASE_URL}/vms/${vm_name}/start")
    
    if echo "$response" | grep -q '"success":true'; then
        log_info "VM started successfully"
        return 0
    else
        log_error "Failed to start VM"
        echo "$response"
        return 1
    fi
}

# Function to get VM details
get_vm_details() {
    local vm_name=$1
    
    log_info "Getting VM details: $vm_name"
    
    response=$(curl -s -X GET "${API_BASE_URL}/vms/${vm_name}")
    
    if echo "$response" | grep -q '"vm"'; then
        echo "$response" | jq . 2>/dev/null || echo "$response"
        return 0
    else
        log_error "Failed to get VM details"
        echo "$response"
        return 1
    fi
}

# Function to create a snapshot
create_snapshot() {
    local vm_name=$1
    local snapshot_name=$2
    
    log_info "Creating snapshot: $snapshot_name for VM: $vm_name"
    
    local snapshot_config=$(cat <<EOF
{
    "name": "${snapshot_name}",
    "description": "Snapshot after initial Windows setup",
    "include_memory": false
}
EOF
)
    
    response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$snapshot_config" \
        "${API_BASE_URL}/vms/${vm_name}/snapshots")
    
    if echo "$response" | grep -q '"snapshot"'; then
        log_info "Snapshot created successfully"
        echo "$response" | jq . 2>/dev/null || echo "$response"
        return 0
    else
        log_error "Failed to create snapshot"
        echo "$response"
        return 1
    fi
}

# Function to list VMs
list_vms() {
    log_info "Listing all VMs..."
    
    response=$(curl -s -X GET "${API_BASE_URL}/vms")
    
    if echo "$response" | grep -q '"vms"'; then
        echo "$response" | jq -r '.vms[] | "\(.name) - \(.status) - CPU: \(.cpu.count) - Memory: \(.memory.size_bytes | tonumber / 1073741824)GB"' 2>/dev/null || echo "$response"
        return 0
    else
        log_error "Failed to list VMs"
        echo "$response"
        return 1
    fi
}

# Main execution
main() {
    log_info "Windows VM Setup Script"
    log_info "API URL: $API_BASE_URL"
    
    # Check if API is available
    if ! check_api_health; then
        log_error "Please ensure the LibGo server is running"
        exit 1
    fi
    
    # Parse command line arguments
    case "${1:-create}" in
        create)
            create_windows_vm "$VM_NAME" "$VM_TEMPLATE"
            ;;
        start)
            start_vm "${2:-$VM_NAME}"
            ;;
        details)
            get_vm_details "${2:-$VM_NAME}"
            ;;
        snapshot)
            create_snapshot "${2:-$VM_NAME}" "${3:-initial-setup}"
            ;;
        list)
            list_vms
            ;;
        full)
            # Full setup: create, start, and snapshot
            create_windows_vm "$VM_NAME" "$VM_TEMPLATE" && \
            sleep 2 && \
            start_vm "$VM_NAME" && \
            log_info "VM is now running. Complete Windows setup, then run: $0 snapshot $VM_NAME post-install"
            ;;
        *)
            echo "Usage: $0 [create|start|details|snapshot|list|full] [vm-name] [snapshot-name]"
            echo ""
            echo "Commands:"
            echo "  create   - Create a new Windows VM (default)"
            echo "  start    - Start an existing VM"
            echo "  details  - Get VM details"
            echo "  snapshot - Create a snapshot"
            echo "  list     - List all VMs"
            echo "  full     - Create and start VM (full setup)"
            echo ""
            echo "Environment variables:"
            echo "  API_BASE_URL - API base URL (default: http://localhost:8080/api/v1)"
            echo "  VM_NAME      - VM name (default: windows11-vm)"
            echo "  VM_TEMPLATE  - VM template (default: windows-11)"
            exit 1
            ;;
    esac
}

# Run main function
main "$@"