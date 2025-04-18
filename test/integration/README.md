# KVM Integration Tests

This directory contains integration tests for the KVM API server.

## Ubuntu Docker Deployment Test

The `ubuntu_docker_test.go` file contains an end-to-end test that:

1. Creates an Ubuntu 24.04 server VM
2. Installs Docker using cloud-init
3. Deploys a Nginx container
4. Creates a custom test page
5. Exports the VM to a compressed file once setup is complete

### Running the Test

To run the Ubuntu Docker deployment test:

```bash
# From the project root
./test/integration/run-docker-test.sh
```

The script will:
- Check if port 8080 is already in use and terminate any process using it
- Start the KVM API server with the test configuration
- Run the integration test
- Report the test results
- Clean up resources

### Test Configuration

The test uses the following configuration:
- VM: Ubuntu 24.04 with 2 vCPUs and 2GB RAM
- Disk: 10GB QCOW2 image
- Network: Default libvirt network
- Docker and Nginx installed via cloud-init

### Authentication

The test uses JWT authentication with the API server. If authentication fails, the test will use a hardcoded token for testing purposes.

### Export

The VM is exported to a QCOW2 file with compression enabled. The export file will be created in the directory specified in the test configuration.
