# Configurable VM Integration Tests

This directory contains a framework for running configurable end-to-end integration tests for the LibGo KVM API. The framework allows for testing various VM deployment scenarios using YAML configuration files, making it easy to create and run tests for different operating systems and configurations.

## Overview

The configurable test framework consists of:

1. **YAML Configuration Files**: Define VM specifications, provisioning methods, verification steps, and export options
2. **Test Runner**: Generic code that executes tests based on the configuration
3. **Test Implementations**: Simple Go files that reference specific configurations

## Directory Structure

- `configtest/`: The framework implementation
  - `loader.go`: YAML configuration file loader
  - `runner.go`: Generic test execution logic
  - `*_test.go`: Individual test implementations
- `test_configs/`: YAML configuration files
  - `ubuntu-docker-test.yaml`: Test configuration for Ubuntu with Docker and Nginx
  - `windows-server-test.yaml`: Test configuration for Windows Server with IIS

## Running Tests

Use the provided script to run tests:

```bash
./run-configurable-integration-test.sh [test-name]
```

Available tests:
- `ubuntu-docker`: Creates an Ubuntu VM, installs Docker and Nginx, and exports it
- `windows-server`: Creates a Windows Server VM, installs IIS, and exports it (requires Windows Server image)

## Creating New Tests

To create a new test:

1. Create a YAML configuration file in `test_configs/`
2. Create a corresponding VM template JSON file in `configs/templates/` if needed
3. Create a test implementation file in `configtest/`
4. Update the `run-configurable-integration-test.sh` script to support the new test

### YAML Configuration Format

```yaml
test:
  name: "Test Name"
  description: "Test description"
  timeout: "90m"  # Test timeout in duration format

vm:
  name: "vm-name"
  template: "template-name"  # References a template JSON file
  description: "VM description"

  # Override template settings if needed
  cpu:
    count: 2
  memory:
    sizeBytes: 2147483648  # 2GB
  disk:
    sizeBytes: 10737418240  # 10GB
    format: "qcow2"

  # OS-specific provisioning
  provisioning:
    method: "cloudinit|unattended"  # cloudinit for Linux, unattended for Windows

    # For Linux
    scripts:
      - name: "Script name"
        content: |
          #cloud-config
          # Cloud-init configuration

    # For Windows
    unattendedXml: |
      <?xml version="1.0" encoding="UTF-8"?>
      <unattend>
        <!-- Windows unattended XML -->
      </unattend>

verification:
  services:
    - name: "Service name"
      port: 80
      protocol: "http"
      expectedContent: "Expected content"
      timeout: 60  # Seconds to wait for provisioning

export:
  format: "qcow2"
  options:
    compress: "true"
    keep_export: "true"
