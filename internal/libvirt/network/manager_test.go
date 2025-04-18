package network

import (
	"context"
	"errors"
	"testing"

	"github.com/digitalocean/go-libvirt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wroersma/libgo/internal/libvirt/connection"
	"github.com/wroersma/libgo/pkg/logger"
)

// MockConnection is a mock implementation of the connection.Connection interface
type MockConnection struct {
	LibvirtConn *MockLibvirtConnection
}

func (m *MockConnection) GetLibvirtConnection() *libvirt.Connect {
	return nil // Not used in tests as we mock the underlying methods
}

func (m *MockConnection) Close() error {
	return nil
}

func (m *MockConnection) IsActive() bool {
	return true
}

// MockLibvirtConnection mocks the libvirt connection interface
type MockLibvirtConnection struct {
	NetworkLookupByNameFunc  func(name string) (libvirt.Network, error)
	NetworkDefineXMLFunc     func(xml string) (libvirt.Network, error)
	NetworkCreateFunc        func(network libvirt.Network) error
	NetworkSetAutostartFunc  func(network libvirt.Network, autostart uint32) error
	NetworkDestroyFunc       func(network libvirt.Network) error
	NetworkUndefineFunc      func(network libvirt.Network) error
	NetworkIsActiveFunc      func(network libvirt.Network) (int32, error)
	NetworkGetDHCPLeasesFunc func(network libvirt.Network, mac *string, needResults uint32, flags uint32) ([]libvirt.NetworkDHCPLease, error)
}

// Test EnsureExists when network already exists
func TestLibvirtNetworkManager_EnsureExists_AlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockLogger(ctrl)
	mockConnManager := connection.NewMockManager(ctrl)
	mockXMLBuilder := NewMockXMLBuilder(ctrl)
	
	mockConn := &MockConnection{
		LibvirtConn: &MockLibvirtConnection{
			NetworkLookupByNameFunc: func(name string) (libvirt.Network, error) {
				return libvirt.Network{}, nil // Return a network, indicating it exists
			},
		},
	}

	// Expect Connect to be called and return our mock connection
	mockConnManager.EXPECT().Connect(gomock.Any()).Return(mockConn, nil)
	mockConnManager.EXPECT().Release(mockConn).Return(nil)

	// Debug log for existing network
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any())

	manager := NewLibvirtNetworkManager(mockConnManager, mockXMLBuilder, mockLogger)
	
	err := manager.EnsureExists(context.Background(), "test-network", "virbr0", "192.168.100.0/24", true)
	assert.NoError(t, err)
}

// Test EnsureExists when network needs to be created
func TestLibvirtNetworkManager_EnsureExists_CreateNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockLogger(ctrl)
	mockConnManager := connection.NewMockManager(ctrl)
	mockXMLBuilder := NewMockXMLBuilder(ctrl)
	
	testNetwork := libvirt.Network{}
	
	mockConn := &MockConnection{
		LibvirtConn: &MockLibvirtConnection{
			// Network doesn't exist
			NetworkLookupByNameFunc: func(name string) (libvirt.Network, error) {
				return libvirt.Network{}, errors.New("network not found")
			},
			// Define network succeeds
			NetworkDefineXMLFunc: func(xml string) (libvirt.Network, error) {
				return testNetwork, nil
			},
			// Create network succeeds
			NetworkCreateFunc: func(network libvirt.Network) error {
				return nil
			},
			// Set autostart succeeds
			NetworkSetAutostartFunc: func(network libvirt.Network, autostart uint32) error {
				return nil
			},
		},
	}

	// Expect Connect to be called and return our mock connection
	mockConnManager.EXPECT().Connect(gomock.Any()).Return(mockConn, nil)
	mockConnManager.EXPECT().Release(mockConn).Return(nil)

	// Expect BuildNetworkXML to be called
	mockXMLBuilder.EXPECT().
		BuildNetworkXML("test-network", "virbr0", "192.168.100.0/24", true).
		Return("<network>...</network>", nil)

	// Debug log for creating network
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())

	manager := NewLibvirtNetworkManager(mockConnManager, mockXMLBuilder, mockLogger)
	
	err := manager.EnsureExists(context.Background(), "test-network", "virbr0", "192.168.100.0/24", true)
	assert.NoError(t, err)
}

// Test Delete when network exists
func TestLibvirtNetworkManager_Delete_Exists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockLogger(ctrl)
	mockConnManager := connection.NewMockManager(ctrl)
	mockXMLBuilder := NewMockXMLBuilder(ctrl)
	
	testNetwork := libvirt.Network{}
	
	mockConn := &MockConnection{
		LibvirtConn: &MockLibvirtConnection{
			// Network exists
			NetworkLookupByNameFunc: func(name string) (libvirt.Network, error) {
				return testNetwork, nil
			},
			// Network is active
			NetworkIsActiveFunc: func(network libvirt.Network) (int32, error) {
				return 1, nil // 1 means active
			},
			// Destroy network succeeds
			NetworkDestroyFunc: func(network libvirt.Network) error {
				return nil
			},
			// Undefine network succeeds
			NetworkUndefineFunc: func(network libvirt.Network) error {
				return nil
			},
		},
	}

	// Expect Connect to be called and return our mock connection
	mockConnManager.EXPECT().Connect(gomock.Any()).Return(mockConn, nil)
	mockConnManager.EXPECT().Release(mockConn).Return(nil)

	// Info log for deleted network
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any())

	manager := NewLibvirtNetworkManager(mockConnManager, mockXMLBuilder, mockLogger)
	
	err := manager.Delete(context.Background(), "test-network")
	assert.NoError(t, err)
}

// Test GetDHCPLeases
func TestLibvirtNetworkManager_GetDHCPLeases(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockLogger(ctrl)
	mockConnManager := connection.NewMockManager(ctrl)
	mockXMLBuilder := NewMockXMLBuilder(ctrl)
	
	testNetwork := libvirt.Network{}
	testLeases := []libvirt.NetworkDHCPLease{
		{
			IPaddr: "192.168.100.101",
			Mac:    "52:54:00:a1:b2:c3",
		},
	}
	
	mockConn := &MockConnection{
		LibvirtConn: &MockLibvirtConnection{
			// Network exists
			NetworkLookupByNameFunc: func(name string) (libvirt.Network, error) {
				return testNetwork, nil
			},
			// Get DHCP leases
			NetworkGetDHCPLeasesFunc: func(network libvirt.Network, mac *string, needResults uint32, flags uint32) ([]libvirt.NetworkDHCPLease, error) {
				return testLeases, nil
			},
		},
	}

	// Expect Connect to be called and return our mock connection
	mockConnManager.EXPECT().Connect(gomock.Any()).Return(mockConn, nil)
	mockConnManager.EXPECT().Release(mockConn).Return(nil)

	manager := NewLibvirtNetworkManager(mockConnManager, mockXMLBuilder, mockLogger)
	
	leases, err := manager.GetDHCPLeases(context.Background(), "test-network")
	require.NoError(t, err)
	assert.Equal(t, testLeases, leases)
}

// Test FindIPByMAC
func TestLibvirtNetworkManager_FindIPByMAC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockLogger(ctrl)
	mockConnManager := connection.NewMockManager(ctrl)
	mockXMLBuilder := NewMockXMLBuilder(ctrl)
	
	testNetwork := libvirt.Network{}
	testLeases := []libvirt.NetworkDHCPLease{
		{
			IPaddr: "192.168.100.101",
			Mac:    "52:54:00:a1:b2:c3",
		},
		{
			IPaddr: "192.168.100.102",
			Mac:    "52:54:00:d4:e5:f6",
		},
	}
	
	mockConn := &MockConnection{
		LibvirtConn: &MockLibvirtConnection{
			// Network exists
			NetworkLookupByNameFunc: func(name string) (libvirt.Network, error) {
				return testNetwork, nil
			},
			// Get DHCP leases
			NetworkGetDHCPLeasesFunc: func(network libvirt.Network, mac *string, needResults uint32, flags uint32) ([]libvirt.NetworkDHCPLease, error) {
				return testLeases, nil
			},
		},
	}

	// Expect Connect to be called and return our mock connection
	mockConnManager.EXPECT().Connect(gomock.Any()).Return(mockConn, nil)
	mockConnManager.EXPECT().Release(mockConn).Return(nil)

	manager := NewLibvirtNetworkManager(mockConnManager, mockXMLBuilder, mockLogger)
	
	// Test finding an existing MAC
	ip, err := manager.FindIPByMAC(context.Background(), "test-network", "52:54:00:d4:e5:f6")
	require.NoError(t, err)
	assert.Equal(t, "192.168.100.102", ip)
	
	// Test with different MAC format (dashes instead of colons)
	mockConnManager.EXPECT().Connect(gomock.Any()).Return(mockConn, nil)
	mockConnManager.EXPECT().Release(mockConn).Return(nil)
	
	ip, err = manager.FindIPByMAC(context.Background(), "test-network", "52-54-00-a1-b2-c3")
	require.NoError(t, err)
	assert.Equal(t, "192.168.100.101", ip)
	
	// Test with MAC not in leases
	mockConnManager.EXPECT().Connect(gomock.Any()).Return(mockConn, nil)
	mockConnManager.EXPECT().Release(mockConn).Return(nil)
	
	_, err = manager.FindIPByMAC(context.Background(), "test-network", "52:54:00:00:00:00")
	assert.Error(t, err)
}

// MockXMLBuilder mocks the XMLBuilder interface
type MockXMLBuilder struct {
	BuildNetworkXMLFunc func(name string, bridgeName string, cidr string, dhcp bool) (string, error)
}

func (m *MockXMLBuilder) BuildNetworkXML(name string, bridgeName string, cidr string, dhcp bool) (string, error) {
	return m.BuildNetworkXMLFunc(name, bridgeName, cidr, dhcp)
}

// NewMockXMLBuilder creates a new mock XMLBuilder controlled by gomock
func NewMockXMLBuilder(ctrl *gomock.Controller) *MockXMLBuilder {
	return &MockXMLBuilder{
		BuildNetworkXMLFunc: func(name string, bridgeName string, cidr string, dhcp bool) (string, error) {
			return "<network>...</network>", nil
		},
	}
}