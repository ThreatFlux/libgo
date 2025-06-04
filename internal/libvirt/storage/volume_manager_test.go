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

// MockPoolManager is a mock for PoolManager
type MockPoolManager struct {
	GetFn          func(ctx context.Context, name string) (*libvirt.StoragePool, error)
	EnsureExistsFn func(ctx context.Context, name string, path string) error
	DeleteFn       func(ctx context.Context, name string) error
	ListFn         func(ctx context.Context) ([]*StoragePoolInfo, error)
	GetInfoFn      func(ctx context.Context, name string) (*StoragePoolInfo, error)
	CreateFn       func(ctx context.Context, params *CreatePoolParams) (*StoragePoolInfo, error)
	StartFn        func(ctx context.Context, name string) error
	StopFn         func(ctx context.Context, name string) error
	RefreshFn      func(ctx context.Context, name string) error
	SetAutostartFn func(ctx context.Context, name string, autostart bool) error
	IsActiveFn     func(ctx context.Context, name string) (bool, error)
	GetXMLFn       func(ctx context.Context, name string) (string, error)
}

func (m *MockPoolManager) Get(ctx context.Context, name string) (*libvirt.StoragePool, error) {
	if m.GetFn != nil {
		return m.GetFn(ctx, name)
	}
	return nil, nil
}

func (m *MockPoolManager) EnsureExists(ctx context.Context, name string, path string) error {
	if m.EnsureExistsFn != nil {
		return m.EnsureExistsFn(ctx, name, path)
	}
	return nil
}

func (m *MockPoolManager) Delete(ctx context.Context, name string) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, name)
	}
	return nil
}

func (m *MockPoolManager) List(ctx context.Context) ([]*StoragePoolInfo, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx)
	}
	return nil, nil
}

func (m *MockPoolManager) GetInfo(ctx context.Context, name string) (*StoragePoolInfo, error) {
	if m.GetInfoFn != nil {
		return m.GetInfoFn(ctx, name)
	}
	return nil, nil
}

func (m *MockPoolManager) Create(ctx context.Context, params *CreatePoolParams) (*StoragePoolInfo, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, params)
	}
	return nil, nil
}

func (m *MockPoolManager) Start(ctx context.Context, name string) error {
	if m.StartFn != nil {
		return m.StartFn(ctx, name)
	}
	return nil
}

func (m *MockPoolManager) Stop(ctx context.Context, name string) error {
	if m.StopFn != nil {
		return m.StopFn(ctx, name)
	}
	return nil
}

func (m *MockPoolManager) Refresh(ctx context.Context, name string) error {
	if m.RefreshFn != nil {
		return m.RefreshFn(ctx, name)
	}
	return nil
}

func (m *MockPoolManager) SetAutostart(ctx context.Context, name string, autostart bool) error {
	if m.SetAutostartFn != nil {
		return m.SetAutostartFn(ctx, name, autostart)
	}
	return nil
}

func (m *MockPoolManager) IsActive(ctx context.Context, name string) (bool, error) {
	if m.IsActiveFn != nil {
		return m.IsActiveFn(ctx, name)
	}
	return false, nil
}

func (m *MockPoolManager) GetXML(ctx context.Context, name string) (string, error) {
	if m.GetXMLFn != nil {
		return m.GetXMLFn(ctx, name)
	}
	return "", nil
}

// MockLibvirtWithVolumes is a mock of libvirt with storage volume operations
type MockLibvirtWithVolumes struct {
	libvirt.Libvirt
}

func TestLibvirtVolumeManager_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping libvirt storage test in short mode")
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConn := mocks_connection.NewMockConnection(ctrl)
	mockConnMgr := mocks_connection.NewMockManager(ctrl)

	// Use mock logger from generated mocks
	mockLog := mocks_logger.NewMockLogger(ctrl)

	// Create mock pool manager
	mockPoolManager := &MockPoolManager{
		GetFn: func(ctx context.Context, name string) (*libvirt.StoragePool, error) {
			if name == "active-pool" {
				return &libvirt.StoragePool{Name: name}, nil
			}
			if name == "inactive-pool" {
				return &libvirt.StoragePool{Name: name}, nil
			}
			return nil, ErrPoolNotFound
		},
	}

	// Create a mock XML builder
	mockXMLBuilder := &MockXMLBuilder{
		BuildStorageVolumeXMLFn: func(volName string, capacityBytes uint64, format string) (string, error) {
			return `<volume><name>test-vol</name><capacity unit="bytes">10737418240</capacity><target><format type="qcow2"/></target></volume>`, nil
		},
	}

	// Set up volume manager
	volumeMgr := NewLibvirtVolumeManager(mockConnMgr, mockPoolManager, mockXMLBuilder, mockLog)

	// Mock libvirt implementation
	mockLibvirt := &MockLibvirtWithVolumes{}

	// Test cases
	testCases := []struct {
		name        string
		poolName    string
		volName     string
		capacity    uint64
		format      string
		volExists   bool
		poolActive  bool
		expectError bool
	}{
		{
			name:        "Create new volume in active pool",
			poolName:    "active-pool",
			volName:     "new-vol",
			capacity:    10 * 1024 * 1024 * 1024, // 10GB
			format:      "qcow2",
			volExists:   false,
			poolActive:  true,
			expectError: false,
		},
		{
			name:        "Volume already exists",
			poolName:    "active-pool",
			volName:     "existing-vol",
			capacity:    5 * 1024 * 1024 * 1024, // 5GB
			format:      "qcow2",
			volExists:   true,
			poolActive:  true,
			expectError: true,
		},
		{
			name:        "Inactive pool",
			poolName:    "inactive-pool",
			volName:     "test-vol",
			capacity:    1 * 1024 * 1024 * 1024, // 1GB
			format:      "raw",
			volExists:   false,
			poolActive:  false,
			expectError: true,
		},
		{
			name:        "Pool not found",
			poolName:    "nonexistent-pool",
			volName:     "test-vol",
			capacity:    1 * 1024 * 1024 * 1024, // 1GB
			format:      "raw",
			volExists:   false,
			poolActive:  false,
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
			err := volumeMgr.Create(context.Background(), tc.poolName, tc.volName, tc.capacity, tc.format)

			if tc.expectError {
				assert.Error(t, err)
				switch {
				case tc.volExists:
					assert.ErrorIs(t, err, ErrVolumeExists)
				case !tc.poolActive && tc.poolName == "inactive-pool":
					assert.ErrorIs(t, err, ErrPoolNotActive)
				case tc.poolName == "nonexistent-pool":
					assert.ErrorContains(t, err, "getting storage pool")
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLibvirtVolumeManager_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping libvirt storage test in short mode")
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConn := mocks_connection.NewMockConnection(ctrl)
	mockConnMgr := mocks_connection.NewMockManager(ctrl)

	// Use mock logger from generated mocks
	mockLog := mocks_logger.NewMockLogger(ctrl)

	// Create mock pool manager
	mockPoolManager := &MockPoolManager{
		GetFn: func(ctx context.Context, name string) (*libvirt.StoragePool, error) {
			if name == "test-pool" {
				return &libvirt.StoragePool{Name: name}, nil
			}
			return nil, ErrPoolNotFound
		},
	}

	// Create a mock XML builder
	mockXMLBuilder := &MockXMLBuilder{}

	// Set up volume manager
	volumeMgr := NewLibvirtVolumeManager(mockConnMgr, mockPoolManager, mockXMLBuilder, mockLog)

	// Mock libvirt implementation
	mockLibvirt := &MockLibvirtWithVolumes{}

	// Test cases
	testCases := []struct {
		name        string
		poolName    string
		volName     string
		volExists   bool
		expectError bool
	}{
		{
			name:        "Delete existing volume",
			poolName:    "test-pool",
			volName:     "existing-vol",
			volExists:   true,
			expectError: false,
		},
		{
			name:        "Delete non-existent volume",
			poolName:    "test-pool",
			volName:     "nonexistent-vol",
			volExists:   false,
			expectError: true,
		},
		{
			name:        "Pool not found",
			poolName:    "nonexistent-pool",
			volName:     "test-vol",
			volExists:   false,
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
			err := volumeMgr.Delete(context.Background(), tc.poolName, tc.volName)

			if tc.expectError {
				assert.Error(t, err)
				if !tc.volExists && tc.poolName == "test-pool" {
					assert.ErrorIs(t, err, ErrVolumeNotFound)
				} else if tc.poolName == "nonexistent-pool" {
					assert.ErrorContains(t, err, "getting storage pool")
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLibvirtVolumeManager_GetPath(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping libvirt storage test in short mode")
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConn := mocks_connection.NewMockConnection(ctrl)
	mockConnMgr := mocks_connection.NewMockManager(ctrl)

	// Use mock logger from generated mocks
	mockLog := mocks_logger.NewMockLogger(ctrl)

	// Create mock pool manager
	mockPoolManager := &MockPoolManager{
		GetFn: func(ctx context.Context, name string) (*libvirt.StoragePool, error) {
			if name == "test-pool" {
				return &libvirt.StoragePool{Name: name}, nil
			}
			return nil, ErrPoolNotFound
		},
	}

	// Create a mock XML builder
	mockXMLBuilder := &MockXMLBuilder{}

	// Set up volume manager
	volumeMgr := NewLibvirtVolumeManager(mockConnMgr, mockPoolManager, mockXMLBuilder, mockLog)

	// Mock libvirt implementation
	mockLibvirt := &MockLibvirtWithVolumes{}

	// Test cases
	testCases := []struct {
		name         string
		poolName     string
		volName      string
		volExists    bool
		expectedPath string
		expectError  bool
	}{
		{
			name:         "Get path for existing volume",
			poolName:     "test-pool",
			volName:      "existing-vol",
			volExists:    true,
			expectedPath: "/var/lib/libvirt/storage/test-pool/existing-vol",
			expectError:  false,
		},
		{
			name:         "Volume not found",
			poolName:     "test-pool",
			volName:      "nonexistent-vol",
			volExists:    false,
			expectedPath: "",
			expectError:  true,
		},
		{
			name:         "Pool not found",
			poolName:     "nonexistent-pool",
			volName:      "test-vol",
			volExists:    false,
			expectedPath: "",
			expectError:  true,
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
			path, err := volumeMgr.GetPath(context.Background(), tc.poolName, tc.volName)

			if tc.expectError {
				assert.Error(t, err)
				assert.Empty(t, path)
				if !tc.volExists && tc.poolName == "test-pool" {
					assert.ErrorIs(t, err, ErrVolumeNotFound)
				} else if tc.poolName == "nonexistent-pool" {
					assert.ErrorContains(t, err, "getting storage pool")
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedPath, path)
			}
		})
	}
}

// Mock libvirt methods
func (m *MockLibvirtWithVolumes) StoragePoolGetInfo(pool libvirt.StoragePool) (libvirt.StoragePoolGetInfoRet, error) {
	if pool.Name == "active-pool" {
		return libvirt.StoragePoolGetInfoRet{
			State: uint8(libvirt.StoragePoolRunning),
		}, nil
	} else if pool.Name == "inactive-pool" {
		return libvirt.StoragePoolGetInfoRet{
			State: uint8(libvirt.StoragePoolInactive),
		}, nil
	}
	return libvirt.StoragePoolGetInfoRet{}, libvirt.Error{Code: uint32(libvirt.ErrNoStoragePool)}
}

func (m *MockLibvirtWithVolumes) StorageVolLookupByName(pool libvirt.StoragePool, name string) (libvirt.StorageVol, error) {
	if name == "existing-vol" {
		return libvirt.StorageVol{
			Name: name,
			Pool: pool.Name,
		}, nil
	}
	return libvirt.StorageVol{}, libvirt.Error{Code: uint32(libvirt.ErrNoStorageVol)}
}

func (m *MockLibvirtWithVolumes) StorageVolCreateXML(pool libvirt.StoragePool, xml string, flags uint32) (libvirt.StorageVol, error) {
	return libvirt.StorageVol{
		Name: "test-vol",
		Pool: pool.Name,
	}, nil
}

func (m *MockLibvirtWithVolumes) StorageVolDelete(vol libvirt.StorageVol, flags uint32) error {
	return nil
}

func (m *MockLibvirtWithVolumes) StorageVolGetPath(vol libvirt.StorageVol) (string, error) {
	return "/var/lib/libvirt/storage/" + vol.Pool + "/" + vol.Name, nil
}

func (m *MockLibvirtWithVolumes) StorageVolResize(vol libvirt.StorageVol, capacity uint64, flags uint32) error {
	return nil
}

func (m *MockLibvirtWithVolumes) StorageVolGetXMLDesc(vol libvirt.StorageVol, flags uint32) (string, error) {
	return `<volume><name>` + vol.Name + `</name></volume>`, nil
}

func (m *MockLibvirtWithVolumes) StorageVolCreateXMLFrom(pool libvirt.StoragePool, xml string, vol libvirt.StorageVol, flags uint32) (libvirt.StorageVol, error) {
	return libvirt.StorageVol{
		Name: "cloned-vol",
		Pool: pool.Name,
	}, nil
}
