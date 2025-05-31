package domain

import (
	"context"
	"testing"

	"github.com/digitalocean/go-libvirt"
	"github.com/stretchr/testify/assert"
	"github.com/threatflux/libgo/internal/models/vm"
	"github.com/threatflux/libgo/pkg/logger"
	"github.com/threatflux/libgo/test/mocks/libvirt"
	"go.uber.org/mock/gomock"
)

const testDomainXML = `
<domain type='kvm'>
  <n>test-vm</n>
  <uuid>12345678-1234-1234-1234-123456789012</uuid>
  <memory unit='KiB'>2097152</memory>
  <currentMemory unit='KiB'>2097152</currentMemory>
  <vcpu placement='static'>2</vcpu>
  <os>
    <type arch='x86_64' machine='q35'>hvm</type>
    <boot dev='hd'/>
  </os>
  <devices>
    <disk type='file' device='disk'>
      <driver name='qemu' type='qcow2'/>
      <source file='/home/vtriple/libgo-storage/test-vm.qcow2'/>
      <target dev='vda' bus='virtio'/>
      <boot order='1'/>
    </disk>
    <interface type='bridge'>
      <source bridge='virbr0'/>
      <mac address='52:54:00:11:22:33'/>
      <model type='virtio'/>
    </interface>
  </devices>
</domain>
`

// Mock XML builder
type mockXMLBuilder struct {
	t       *testing.T
	buildFn func(params vm.VMParams) (string, error)
}

func (m *mockXMLBuilder) BuildDomainXML(params vm.VMParams) (string, error) {
	return m.buildFn(params)
}

// Mock libvirt connection
type mockLibvirtWithDomain struct {
	libvirt.Libvirt
	t *testing.T
}

func TestDomainManager_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConn := mocks_libvirt.NewMockConnection(ctrl)
	mockConnMgr := mocks_libvirt.NewMockManager(ctrl)

	// Mock logger
	mockLog := logger.NewZapLogger(LoggingConfig{
		Level:      "debug",
		Format:     "json",
		OutputPath: "stdout",
	})

	// Create a mock XML builder
	xmlBuilder := &mockXMLBuilder{
		t: t,
		buildFn: func(params vm.VMParams) (string, error) {
			assert.Equal(t, "test-vm", params.Name)
			return testDomainXML, nil
		},
	}

	// Set up domain manager
	domainMgr := NewDomainManager(mockConnMgr, xmlBuilder, mockLog)

	// Mock expectations
	mockLibvirt := &mockLibvirtWithDomain{t: t}
	mockConn.EXPECT().GetLibvirtConnection().Return(mockLibvirt).AnyTimes()
	mockConn.EXPECT().IsActive().Return(true).AnyTimes()
	mockConnMgr.EXPECT().Connect(gomock.Any()).Return(mockConn, nil)
	mockConnMgr.EXPECT().Release(mockConn).Return(nil)

	// Create params
	params := vm.VMParams{
		Name: "test-vm",
		CPU: vm.CPUParams{
			Count: 2,
		},
		Memory: vm.MemoryParams{
			SizeBytes: 2 * 1024 * 1024 * 1024, // 2GB
		},
		Disk: vm.DiskParams{
			Format: "qcow2",
		},
		Network: vm.NetParams{
			Type:   "bridge",
			Source: "virbr0",
		},
	}

	// Call Create
	vm, err := domainMgr.Create(context.Background(), params)
	assert.NoError(t, err)
	assert.NotNil(t, vm)
	assert.Equal(t, "test-vm", vm.Name)
}

func TestDomainManager_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConn := mocks_libvirt.NewMockConnection(ctrl)
	mockConnMgr := mocks_libvirt.NewMockManager(ctrl)

	// Mock logger
	mockLog := logger.NewZapLogger(LoggingConfig{
		Level:      "debug",
		Format:     "json",
		OutputPath: "stdout",
	})

	// Create a mock XML builder
	xmlBuilder := &mockXMLBuilder{t: t}

	// Set up domain manager
	domainMgr := NewDomainManager(mockConnMgr, xmlBuilder, mockLog)

	// Mock expectations
	mockLibvirt := &mockLibvirtWithDomain{t: t}
	mockConn.EXPECT().GetLibvirtConnection().Return(mockLibvirt).AnyTimes()
	mockConn.EXPECT().IsActive().Return(true).AnyTimes()
	mockConnMgr.EXPECT().Connect(gomock.Any()).Return(mockConn, nil)
	mockConnMgr.EXPECT().Release(mockConn).Return(nil)

	// Call Get
	vm, err := domainMgr.Get(context.Background(), "test-vm")
	assert.NoError(t, err)
	assert.NotNil(t, vm)
	assert.Equal(t, "test-vm", vm.Name)
}

func TestDomainManager_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConn := mocks_libvirt.NewMockConnection(ctrl)
	mockConnMgr := mocks_libvirt.NewMockManager(ctrl)

	// Mock logger
	mockLog := logger.NewZapLogger(LoggingConfig{
		Level:      "debug",
		Format:     "json",
		OutputPath: "stdout",
	})

	// Create a mock XML builder
	xmlBuilder := &mockXMLBuilder{t: t}

	// Set up domain manager
	domainMgr := NewDomainManager(mockConnMgr, xmlBuilder, mockLog)

	// Mock expectations
	mockLibvirt := &mockLibvirtWithDomain{t: t}
	mockConn.EXPECT().GetLibvirtConnection().Return(mockLibvirt).AnyTimes()
	mockConn.EXPECT().IsActive().Return(true).AnyTimes()
	mockConnMgr.EXPECT().Connect(gomock.Any()).Return(mockConn, nil)
	mockConnMgr.EXPECT().Release(mockConn).Return(nil)

	// Call List
	vms, err := domainMgr.List(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, vms)
	assert.Len(t, vms, 1) // Assuming 1 domain in the test data
}

func TestDomainManager_Start(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConn := mocks_libvirt.NewMockConnection(ctrl)
	mockConnMgr := mocks_libvirt.NewMockManager(ctrl)

	// Mock logger
	mockLog := logger.NewZapLogger(LoggingConfig{
		Level:      "debug",
		Format:     "json",
		OutputPath: "stdout",
	})

	// Create a mock XML builder
	xmlBuilder := &mockXMLBuilder{t: t}

	// Set up domain manager
	domainMgr := NewDomainManager(mockConnMgr, xmlBuilder, mockLog)

	// Mock expectations
	mockLibvirt := &mockLibvirtWithDomain{t: t}
	mockConn.EXPECT().GetLibvirtConnection().Return(mockLibvirt).AnyTimes()
	mockConn.EXPECT().IsActive().Return(true).AnyTimes()
	mockConnMgr.EXPECT().Connect(gomock.Any()).Return(mockConn, nil)
	mockConnMgr.EXPECT().Release(mockConn).Return(nil)

	// Call Start
	err := domainMgr.Start(context.Background(), "test-vm")
	assert.NoError(t, err)
}

func TestDomainManager_Stop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConn := mocks_libvirt.NewMockConnection(ctrl)
	mockConnMgr := mocks_libvirt.NewMockManager(ctrl)

	// Mock logger
	mockLog := logger.NewZapLogger(LoggingConfig{
		Level:      "debug",
		Format:     "json",
		OutputPath: "stdout",
	})

	// Create a mock XML builder
	xmlBuilder := &mockXMLBuilder{t: t}

	// Set up domain manager
	domainMgr := NewDomainManager(mockConnMgr, xmlBuilder, mockLog)

	// Mock expectations
	mockLibvirt := &mockLibvirtWithDomain{t: t}
	mockConn.EXPECT().GetLibvirtConnection().Return(mockLibvirt).AnyTimes()
	mockConn.EXPECT().IsActive().Return(true).AnyTimes()
	mockConnMgr.EXPECT().Connect(gomock.Any()).Return(mockConn, nil)
	mockConnMgr.EXPECT().Release(mockConn).Return(nil)

	// Call Stop
	err := domainMgr.Stop(context.Background(), "test-vm")
	assert.NoError(t, err)
}

func TestDomainManager_ForceStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConn := mocks_libvirt.NewMockConnection(ctrl)
	mockConnMgr := mocks_libvirt.NewMockManager(ctrl)

	// Mock logger
	mockLog := logger.NewZapLogger(LoggingConfig{
		Level:      "debug",
		Format:     "json",
		OutputPath: "stdout",
	})

	// Create a mock XML builder
	xmlBuilder := &mockXMLBuilder{t: t}

	// Set up domain manager
	domainMgr := NewDomainManager(mockConnMgr, xmlBuilder, mockLog)

	// Mock expectations
	mockLibvirt := &mockLibvirtWithDomain{t: t}
	mockConn.EXPECT().GetLibvirtConnection().Return(mockLibvirt).AnyTimes()
	mockConn.EXPECT().IsActive().Return(true).AnyTimes()
	mockConnMgr.EXPECT().Connect(gomock.Any()).Return(mockConn, nil)
	mockConnMgr.EXPECT().Release(mockConn).Return(nil)

	// Call ForceStop
	err := domainMgr.ForceStop(context.Background(), "test-vm")
	assert.NoError(t, err)
}

func TestDomainManager_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConn := mocks_libvirt.NewMockConnection(ctrl)
	mockConnMgr := mocks_libvirt.NewMockManager(ctrl)

	// Mock logger
	mockLog := logger.NewZapLogger(LoggingConfig{
		Level:      "debug",
		Format:     "json",
		OutputPath: "stdout",
	})

	// Create a mock XML builder
	xmlBuilder := &mockXMLBuilder{t: t}

	// Set up domain manager
	domainMgr := NewDomainManager(mockConnMgr, xmlBuilder, mockLog)

	// Mock expectations
	mockLibvirt := &mockLibvirtWithDomain{t: t}
	mockConn.EXPECT().GetLibvirtConnection().Return(mockLibvirt).AnyTimes()
	mockConn.EXPECT().IsActive().Return(true).AnyTimes()
	mockConnMgr.EXPECT().Connect(gomock.Any()).Return(mockConn, nil)
	mockConnMgr.EXPECT().Release(mockConn).Return(nil)

	// Call Delete
	err := domainMgr.Delete(context.Background(), "test-vm")
	assert.NoError(t, err)
}

func TestDomainManager_GetXML(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConn := mocks_libvirt.NewMockConnection(ctrl)
	mockConnMgr := mocks_libvirt.NewMockManager(ctrl)

	// Mock logger
	mockLog := logger.NewZapLogger(LoggingConfig{
		Level:      "debug",
		Format:     "json",
		OutputPath: "stdout",
	})

	// Create a mock XML builder
	xmlBuilder := &mockXMLBuilder{t: t}

	// Set up domain manager
	domainMgr := NewDomainManager(mockConnMgr, xmlBuilder, mockLog)

	// Mock expectations
	mockLibvirt := &mockLibvirtWithDomain{t: t}
	mockConn.EXPECT().GetLibvirtConnection().Return(mockLibvirt).AnyTimes()
	mockConn.EXPECT().IsActive().Return(true).AnyTimes()
	mockConnMgr.EXPECT().Connect(gomock.Any()).Return(mockConn, nil)
	mockConnMgr.EXPECT().Release(mockConn).Return(nil)

	// Call GetXML
	xml, err := domainMgr.GetXML(context.Background(), "test-vm")
	assert.NoError(t, err)
	assert.Equal(t, testDomainXML, xml)
}

// Mock libvirt methods
func (m *mockLibvirtWithDomain) DomainLookupByName(name string) (libvirt.Domain, error) {
	return libvirt.Domain{
		Name: name,
	}, nil
}

func (m *mockLibvirtWithDomain) DomainGetXMLDesc(domain libvirt.Domain, flags uint32) (string, error) {
	return testDomainXML, nil
}

func (m *mockLibvirtWithDomain) DomainGetInfo(domain libvirt.Domain) (libvirt.DomainInfo, error) {
	return libvirt.DomainInfo{
		State:     libvirt.DomainRunning,
		MaxMem:    2097152, // 2GB in KiB
		Memory:    2097152,
		NrVirtCpu: 2,
	}, nil
}

func (m *mockLibvirtWithDomain) DomainCreate(domain libvirt.Domain) error {
	return nil
}

func (m *mockLibvirtWithDomain) DomainShutdown(domain libvirt.Domain) error {
	return nil
}

func (m *mockLibvirtWithDomain) DomainDestroy(domain libvirt.Domain) error {
	return nil
}

func (m *mockLibvirtWithDomain) DomainUndefine(domain libvirt.Domain) error {
	return nil
}

func (m *mockLibvirtWithDomain) DomainUndefineFlags(domain libvirt.Domain, flags uint32) error {
	return nil
}

func (m *mockLibvirtWithDomain) DomainDefineXML(xmlConfig string) (libvirt.Domain, error) {
	return libvirt.Domain{
		Name: "test-vm",
	}, nil
}

func (m *mockLibvirtWithDomain) Domains() ([]libvirt.Domain, error) {
	return []libvirt.Domain{
		{
			Name: "test-vm",
		},
	}, nil
}

// LoggingConfig for tests
type LoggingConfig struct {
	Level      string
	Format     string
	OutputPath string
}
