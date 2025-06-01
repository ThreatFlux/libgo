#!/usr/bin/env python3
"""
Windows VM Creation Script using LibGo API
This script demonstrates creating multiple Windows VMs with different configurations
"""

import json
import time
import requests
from typing import Dict, List, Optional, Tuple
import argparse
import sys

# Configuration
API_BASE_URL = "http://localhost:8080/api/v1"
DEFAULT_TIMEOUT = 30

# VM Templates
VM_TEMPLATES = {
    "windows-11-basic": {
        "template": "windows-11",
        "cpu": {"count": 4, "cores": 2, "threads": 1, "socket": 2},
        "memory": {"sizeBytes": 8 * 1024 * 1024 * 1024},  # 8GB
        "disk": {"sizeBytes": 100 * 1024 * 1024 * 1024, "format": "qcow2", "bus": "sata"}
    },
    "windows-11-developer": {
        "template": "windows-11",
        "cpu": {"count": 8, "cores": 4, "threads": 1, "socket": 2},
        "memory": {"sizeBytes": 16 * 1024 * 1024 * 1024},  # 16GB
        "disk": {"sizeBytes": 250 * 1024 * 1024 * 1024, "format": "qcow2", "bus": "virtio"}
    },
    "windows-server-web": {
        "template": "windows-server-2022",
        "cpu": {"count": 4, "cores": 2, "threads": 1, "socket": 2},
        "memory": {"sizeBytes": 8 * 1024 * 1024 * 1024},  # 8GB
        "disk": {"sizeBytes": 80 * 1024 * 1024 * 1024, "format": "qcow2", "bus": "virtio"}
    },
    "windows-server-database": {
        "template": "windows-server-2022",
        "cpu": {"count": 8, "cores": 4, "threads": 1, "socket": 2},
        "memory": {"sizeBytes": 32 * 1024 * 1024 * 1024},  # 32GB
        "disk": {"sizeBytes": 500 * 1024 * 1024 * 1024, "format": "qcow2", "bus": "virtio"}
    }
}

class LibGoClient:
    """Client for interacting with LibGo API"""
    
    def __init__(self, base_url: str = API_BASE_URL, auth_token: Optional[str] = None):
        self.base_url = base_url.rstrip('/')
        self.session = requests.Session()
        if auth_token:
            self.session.headers['Authorization'] = f'Bearer {auth_token}'
    
    def check_health(self) -> bool:
        """Check if API is healthy"""
        try:
            response = self.session.get(
                f"{self.base_url.replace('/api/v1', '')}/health",
                timeout=DEFAULT_TIMEOUT
            )
            return response.status_code == 200
        except Exception as e:
            print(f"Health check failed: {e}")
            return False
    
    def create_vm(self, vm_config: Dict) -> Tuple[bool, Dict]:
        """Create a VM with given configuration"""
        try:
            response = self.session.post(
                f"{self.base_url}/vms",
                json=vm_config,
                timeout=DEFAULT_TIMEOUT
            )
            
            if response.status_code == 201:
                return True, response.json()
            else:
                return False, {"error": response.text, "status_code": response.status_code}
        except Exception as e:
            return False, {"error": str(e)}
    
    def start_vm(self, vm_name: str) -> Tuple[bool, Dict]:
        """Start a VM"""
        try:
            response = self.session.put(
                f"{self.base_url}/vms/{vm_name}/start",
                timeout=DEFAULT_TIMEOUT
            )
            
            if response.status_code == 200:
                return True, response.json()
            else:
                return False, {"error": response.text, "status_code": response.status_code}
        except Exception as e:
            return False, {"error": str(e)}
    
    def get_vm(self, vm_name: str) -> Tuple[bool, Dict]:
        """Get VM details"""
        try:
            response = self.session.get(
                f"{self.base_url}/vms/{vm_name}",
                timeout=DEFAULT_TIMEOUT
            )
            
            if response.status_code == 200:
                return True, response.json()
            else:
                return False, {"error": response.text, "status_code": response.status_code}
        except Exception as e:
            return False, {"error": str(e)}
    
    def list_vms(self) -> Tuple[bool, List[Dict]]:
        """List all VMs"""
        try:
            response = self.session.get(
                f"{self.base_url}/vms",
                timeout=DEFAULT_TIMEOUT
            )
            
            if response.status_code == 200:
                data = response.json()
                return True, data.get('vms', [])
            else:
                return False, []
        except Exception as e:
            print(f"List VMs failed: {e}")
            return False, []
    
    def create_snapshot(self, vm_name: str, snapshot_name: str, 
                       description: str = "", include_memory: bool = False) -> Tuple[bool, Dict]:
        """Create a VM snapshot"""
        snapshot_config = {
            "name": snapshot_name,
            "description": description,
            "include_memory": include_memory
        }
        
        try:
            response = self.session.post(
                f"{self.base_url}/vms/{vm_name}/snapshots",
                json=snapshot_config,
                timeout=DEFAULT_TIMEOUT
            )
            
            if response.status_code == 201:
                return True, response.json()
            else:
                return False, {"error": response.text, "status_code": response.status_code}
        except Exception as e:
            return False, {"error": str(e)}
    
    def delete_vm(self, vm_name: str) -> Tuple[bool, Dict]:
        """Delete a VM"""
        try:
            response = self.session.delete(
                f"{self.base_url}/vms/{vm_name}",
                timeout=DEFAULT_TIMEOUT
            )
            
            if response.status_code in [200, 204]:
                return True, {"message": "VM deleted successfully"}
            else:
                return False, {"error": response.text, "status_code": response.status_code}
        except Exception as e:
            return False, {"error": str(e)}

def create_windows_vm(client: LibGoClient, name: str, template_type: str, 
                     description: str = "") -> bool:
    """Create a Windows VM using predefined template"""
    
    if template_type not in VM_TEMPLATES:
        print(f"Unknown template type: {template_type}")
        print(f"Available templates: {', '.join(VM_TEMPLATES.keys())}")
        return False
    
    template = VM_TEMPLATES[template_type]
    
    vm_config = {
        "name": name,
        "description": description or f"Windows VM - {template_type}",
        **template,
        "network": {
            "type": "network",
            "source": "default",
            "model": "virtio"
        },
        "cloudInit": {
            "enabled": False
        }
    }
    
    print(f"\nCreating VM: {name}")
    print(f"Template: {template_type}")
    print(f"CPU: {template['cpu']['count']} cores")
    print(f"Memory: {template['memory']['sizeBytes'] / (1024**3):.0f} GB")
    print(f"Disk: {template['disk']['sizeBytes'] / (1024**3):.0f} GB")
    
    success, result = client.create_vm(vm_config)
    
    if success:
        vm = result.get('vm', {})
        print(f"✓ VM created successfully!")
        print(f"  UUID: {vm.get('uuid')}")
        print(f"  Status: {vm.get('status')}")
        return True
    else:
        print(f"✗ Failed to create VM: {result}")
        return False

def main():
    parser = argparse.ArgumentParser(description='Create Windows VMs using LibGo API')
    parser.add_argument('--api-url', default=API_BASE_URL, help='API base URL')
    parser.add_argument('--token', help='Authentication token')
    
    subparsers = parser.add_subparsers(dest='command', help='Commands')
    
    # Create command
    create_parser = subparsers.add_parser('create', help='Create a VM')
    create_parser.add_argument('name', help='VM name')
    create_parser.add_argument('--template', default='windows-11-basic', 
                             choices=list(VM_TEMPLATES.keys()),
                             help='VM template to use')
    create_parser.add_argument('--description', help='VM description')
    create_parser.add_argument('--start', action='store_true', help='Start VM after creation')
    
    # Create multiple command
    multi_parser = subparsers.add_parser('create-multiple', help='Create multiple VMs')
    multi_parser.add_argument('--config-file', help='JSON file with VM configurations')
    
    # List command
    list_parser = subparsers.add_parser('list', help='List all VMs')
    
    # Delete command
    delete_parser = subparsers.add_parser('delete', help='Delete a VM')
    delete_parser.add_argument('name', help='VM name')
    
    # Start command
    start_parser = subparsers.add_parser('start', help='Start a VM')
    start_parser.add_argument('name', help='VM name')
    
    # Snapshot command
    snapshot_parser = subparsers.add_parser('snapshot', help='Create a snapshot')
    snapshot_parser.add_argument('vm_name', help='VM name')
    snapshot_parser.add_argument('snapshot_name', help='Snapshot name')
    snapshot_parser.add_argument('--description', help='Snapshot description')
    snapshot_parser.add_argument('--include-memory', action='store_true', 
                               help='Include memory state')
    
    args = parser.parse_args()
    
    # Create client
    client = LibGoClient(args.api_url, args.token)
    
    # Check API health
    if not client.check_health():
        print("Error: API is not responding. Please check if LibGo server is running.")
        sys.exit(1)
    
    # Execute command
    if args.command == 'create':
        success = create_windows_vm(client, args.name, args.template, args.description)
        
        if success and args.start:
            print(f"\nStarting VM: {args.name}")
            success, result = client.start_vm(args.name)
            if success:
                print("✓ VM started successfully!")
            else:
                print(f"✗ Failed to start VM: {result}")
    
    elif args.command == 'create-multiple':
        if args.config_file:
            # Load configurations from file
            with open(args.config_file, 'r') as f:
                configs = json.load(f)
            
            for config in configs:
                create_windows_vm(client, config['name'], config['template'], 
                                config.get('description', ''))
                time.sleep(2)  # Small delay between creations
        else:
            # Create default set of VMs
            vms = [
                ("win11-workstation", "windows-11-basic", "Windows 11 Workstation"),
                ("win11-dev", "windows-11-developer", "Windows 11 Development"),
                ("winserver-web", "windows-server-web", "Windows Server - Web Server"),
                ("winserver-db", "windows-server-database", "Windows Server - Database")
            ]
            
            for name, template, description in vms:
                create_windows_vm(client, name, template, description)
                time.sleep(2)
    
    elif args.command == 'list':
        success, vms = client.list_vms()
        if success:
            if vms:
                print(f"\nFound {len(vms)} VMs:")
                print("-" * 80)
                for vm in vms:
                    print(f"Name: {vm['name']}")
                    print(f"  Status: {vm['status']}")
                    print(f"  CPU: {vm['cpu']['count']} cores")
                    print(f"  Memory: {vm['memory']['size_bytes'] / (1024**3):.1f} GB")
                    print(f"  Created: {vm['created_at']}")
                    print("-" * 80)
            else:
                print("No VMs found.")
        else:
            print("Failed to list VMs")
    
    elif args.command == 'delete':
        print(f"Deleting VM: {args.name}")
        success, result = client.delete_vm(args.name)
        if success:
            print("✓ VM deleted successfully!")
        else:
            print(f"✗ Failed to delete VM: {result}")
    
    elif args.command == 'start':
        print(f"Starting VM: {args.name}")
        success, result = client.start_vm(args.name)
        if success:
            print("✓ VM started successfully!")
        else:
            print(f"✗ Failed to start VM: {result}")
    
    elif args.command == 'snapshot':
        print(f"Creating snapshot for VM: {args.vm_name}")
        success, result = client.create_snapshot(
            args.vm_name, 
            args.snapshot_name,
            args.description or f"Snapshot of {args.vm_name}",
            args.include_memory
        )
        if success:
            snapshot = result.get('snapshot', {})
            print("✓ Snapshot created successfully!")
            print(f"  Name: {snapshot.get('name')}")
            print(f"  Created: {snapshot.get('created_at')}")
        else:
            print(f"✗ Failed to create snapshot: {result}")
    
    else:
        parser.print_help()

if __name__ == "__main__":
    main()