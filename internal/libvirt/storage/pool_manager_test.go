package storage

import (
	"context"
	"testing"

	"github.com/digitalocean/go-libvirt"
	"github.com/stretchr/testify/assert"
	mocks_connection "github.com/threatflux/libgo/test/mocks/libvirt/connection"
	mocks_logger "github.com/threatflux/libgo/test/mocks/logger"
	"go.uber.org/mock/gomock"
)

// MockXMLBuilder is a mock for storage XML builder
type MockXMLBuilder struct {
	BuildStoragePoolXMLFn   func(name string, path string) (string, error)
	BuildStorageVolumeXMLFn func(volName string, capacityBytes uint64, format string) (string, error)
}

func (m *MockXMLBuilder) BuildStoragePoolXML(name string, path string) (string, error) {
	return m.BuildStoragePoolXMLFn(name, path)
}

func (m *MockXMLBuilder) BuildStorageVolumeXML(volName string, capacityBytes uint64, format string) (string, error) {
	return m.BuildStorageVolumeXMLFn(volName, capacityBytes, format)
}

// MockLibvirtWithPools is a mock of libvirt with storage pool operations
type MockLibvirtWithPools struct {
	libvirt.Libvirt
}

func TestLibvirtPoolManager_EnsureExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConn := mocks_connection.NewMockConnection(ctrl)
	mockConnMgr := mocks_connection.NewMockManager(ctrl)

	// Use mock logger from generated mocks
	mockLog := mocks_logger.NewMockLogger(ctrl)

	// Create a mock XML builder
	mockXMLBuilder := &MockXMLBuilder{
		BuildStoragePoolXMLFn: func(name string, path string) (string, error) {
			return `<pool type='dir'><n>test-pool</n><target><path>/var/lib/libvirt/storage/test-pool</path></target></pool>`, nil
		},
	}

	// Set up pool manager
	poolMgr := NewLibvirtPoolManager(mockConnMgr, mockXMLBuilder, mockLog)

	// Mock libvirt implementation
	mockLibvirt := &MockLibvirtWithPools{}

	// Test cases
	testCases := []struct {
		name        string
		poolName    string
		poolPath    string
		poolExists  bool
		poolRunning bool
		expectError bool
	}{
		{
			name:        "New pool creation",
			poolName:    "test-pool",
			poolPath:    "/var/lib/libvirt/storage/test-pool",
			poolExists:  false,
			poolRunning: false,
			expectError: false,
		},
		{
			name:        "Existing and running pool",
			poolName:    "existing-pool",
			poolPath:    "/var/lib/libvirt/storage/existing-pool",
			poolExists:  true,
			poolRunning: true,
			expectError: false,
		},
		{
			name:        "Existing but inactive pool",
			poolName:    "inactive-pool",
			poolPath:    "/var/lib/libvirt/storage/inactive-pool",
			poolExists:  true,
			poolRunning: false,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Mock connection behavior
			mockConn.EXPECT().GetLibvirtConnection().Return(mockLibvirt).AnyTimes()
			mockConn.EXPECT().IsActive().Return(true).AnyTimes()
			mockConnMgr.EXPECT().Connect(gomock.Any()).Return(mockConn, nil)
			mockConnMgr.EXPECT().Release(mockConn).Return(nil)

			// Run the test
			err := poolMgr.EnsureExists(context.Background(), tc.poolName, tc.poolPath)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLibvirtPoolManager_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConn := mocks_connection.NewMockConnection(ctrl)
	mockConnMgr := mocks_connection.NewMockManager(ctrl)

	// Use mock logger from generated mocks
	mockLog := mocks_logger.NewMockLogger(ctrl)

	// Create a mock XML builder
	mockXMLBuilder := &MockXMLBuilder{}

	// Set up pool manager
	poolMgr := NewLibvirtPoolManager(mockConnMgr, mockXMLBuilder, mockLog)

	// Mock libvirt implementation
	mockLibvirt := &MockLibvirtWithPools{}

	// Test cases
	testCases := []struct {
		name        string
		poolName    string
		poolExists  bool
		poolRunning bool
		expectError bool
	}{
		{
			name:        "Delete existing running pool",
			poolName:    "running-pool",
			poolExists:  true,
			poolRunning: true,
			expectError: false,
		},
		{
			name:        "Delete existing inactive pool",
			poolName:    "inactive-pool",
			poolExists:  true,
			poolRunning: false,
			expectError: false,
		},
		{
			name:        "Delete non-existent pool",
			poolName:    "nonexistent-pool",
			poolExists:  false,
			poolRunning: false,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Mock connection behavior
			mockConn.EXPECT().GetLibvirtConnection().Return(mockLibvirt).AnyTimes()
			mockConn.EXPECT().IsActive().Return(true).AnyTimes()
			mockConnMgr.EXPECT().Connect(gomock.Any()).Return(mockConn, nil)
			mockConnMgr.EXPECT().Release(mockConn).Return(nil)

			// Run the test
			err := poolMgr.Delete(context.Background(), tc.poolName)

			if tc.expectError {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrPoolNotFound)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLibvirtPoolManager_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConn := mocks_connection.NewMockConnection(ctrl)
	mockConnMgr := mocks_connection.NewMockManager(ctrl)

	// Use mock logger from generated mocks
	mockLog := mocks_logger.NewMockLogger(ctrl)

	// Create a mock XML builder
	mockXMLBuilder := &MockXMLBuilder{}

	// Set up pool manager
	poolMgr := NewLibvirtPoolManager(mockConnMgr, mockXMLBuilder, mockLog)

	// Mock libvirt implementation
	mockLibvirt := &MockLibvirtWithPools{}

	// Test cases
	testCases := []struct {
		name        string
		poolName    string
		poolExists  bool
		expectError bool
	}{
		{
			name:        "Get existing pool",
			poolName:    "existing-pool",
			poolExists:  true,
			expectError: false,
		},
		{
			name:        "Get non-existent pool",
			poolName:    "nonexistent-pool",
			poolExists:  false,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Mock connection behavior
			mockConn.EXPECT().GetLibvirtConnection().Return(mockLibvirt).AnyTimes()
			mockConn.EXPECT().IsActive().Return(true).AnyTimes()
			mockConnMgr.EXPECT().Connect(gomock.Any()).Return(mockConn, nil)
			mockConnMgr.EXPECT().Release(mockConn).Return(nil)

			// Run the test
			pool, err := poolMgr.Get(context.Background(), tc.poolName)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, pool)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, pool)
				assert.Equal(t, tc.poolName, pool.Name)
			}
		})
	}
}

// LoggingConfig for tests
type LoggingConfig struct {
	Level      string
	Format     string
	OutputPath string
}

// Mock libvirt methods
func (m *MockLibvirtWithPools) StoragePoolLookupByName(name string) (libvirt.StoragePool, error) {
	switch name {
	case "existing-pool", "running-pool":
		return libvirt.StoragePool{
			Name: name,
		}, nil
	case "inactive-pool":
		return libvirt.StoragePool{
			Name: name,
		}, nil
	default:
		return libvirt.StoragePool{}, libvirt.Error{Code: uint32(libvirt.ErrNoStoragePool)}
	}
}

func (m *MockLibvirtWithPools) StoragePoolGetInfo(pool libvirt.StoragePool) (libvirt.StoragePoolGetInfoRet, error) {
	switch pool.Name {
	case "existing-pool", "running-pool":
		return libvirt.StoragePoolGetInfoRet{
			State: uint8(libvirt.StoragePoolRunning),
		}, nil
	case "inactive-pool":
		return libvirt.StoragePoolGetInfoRet{
			State: uint8(libvirt.StoragePoolInactive),
		}, nil
	default:
		return libvirt.StoragePoolGetInfoRet{}, libvirt.Error{Code: uint32(libvirt.ErrNoStoragePool)}
	}
}

func (m *MockLibvirtWithPools) StoragePoolCreate(pool libvirt.StoragePool) error {
	return nil
}

func (m *MockLibvirtWithPools) StoragePoolDestroy(pool libvirt.StoragePool) error {
	return nil
}

func (m *MockLibvirtWithPools) StoragePoolUndefine(pool libvirt.StoragePool) error {
	return nil
}

func (m *MockLibvirtWithPools) StoragePoolDefineXML(xml string) (libvirt.StoragePool, error) {
	return libvirt.StoragePool{
		Name: "test-pool",
	}, nil
}

func (m *MockLibvirtWithPools) StoragePoolSetAutostart(pool libvirt.StoragePool, autostart uint32) error {
	return nil
}
