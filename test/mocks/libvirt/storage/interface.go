// Code generated by MockGen. DO NOT EDIT.
// Source: internal/libvirt/storage/interface.go
//
// Generated by this command:
//
//	mockgen -source=internal/libvirt/storage/interface.go -destination=./test/mocks/libvirt/storage/interface.go -package=mocks_storage
//

// Package mocks_storage is a generated GoMock package.
package mocks_storage

import (
	context "context"
	io "io"
	reflect "reflect"

	libvirt "github.com/digitalocean/go-libvirt"
	storage "github.com/threatflux/libgo/internal/libvirt/storage"
	gomock "go.uber.org/mock/gomock"
)

// MockPoolManager is a mock of PoolManager interface.
type MockPoolManager struct {
	isgomock struct{}
	ctrl     *gomock.Controller
	recorder *MockPoolManagerMockRecorder
}

// MockPoolManagerMockRecorder is the mock recorder for MockPoolManager.
type MockPoolManagerMockRecorder struct {
	mock *MockPoolManager
}

// NewMockPoolManager creates a new mock instance.
func NewMockPoolManager(ctrl *gomock.Controller) *MockPoolManager {
	mock := &MockPoolManager{ctrl: ctrl}
	mock.recorder = &MockPoolManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPoolManager) EXPECT() *MockPoolManagerMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockPoolManager) Create(ctx context.Context, params *storage.CreatePoolParams) (*storage.StoragePoolInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, params)
	ret0, _ := ret[0].(*storage.StoragePoolInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *MockPoolManagerMockRecorder) Create(ctx, params any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockPoolManager)(nil).Create), ctx, params)
}

// Delete mocks base method.
func (m *MockPoolManager) Delete(ctx context.Context, name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, name)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockPoolManagerMockRecorder) Delete(ctx, name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockPoolManager)(nil).Delete), ctx, name)
}

// EnsureExists mocks base method.
func (m *MockPoolManager) EnsureExists(ctx context.Context, name, path string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnsureExists", ctx, name, path)
	ret0, _ := ret[0].(error)
	return ret0
}

// EnsureExists indicates an expected call of EnsureExists.
func (mr *MockPoolManagerMockRecorder) EnsureExists(ctx, name, path any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnsureExists", reflect.TypeOf((*MockPoolManager)(nil).EnsureExists), ctx, name, path)
}

// Get mocks base method.
func (m *MockPoolManager) Get(ctx context.Context, name string) (*libvirt.StoragePool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, name)
	ret0, _ := ret[0].(*libvirt.StoragePool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockPoolManagerMockRecorder) Get(ctx, name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockPoolManager)(nil).Get), ctx, name)
}

// GetInfo mocks base method.
func (m *MockPoolManager) GetInfo(ctx context.Context, name string) (*storage.StoragePoolInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetInfo", ctx, name)
	ret0, _ := ret[0].(*storage.StoragePoolInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetInfo indicates an expected call of GetInfo.
func (mr *MockPoolManagerMockRecorder) GetInfo(ctx, name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInfo", reflect.TypeOf((*MockPoolManager)(nil).GetInfo), ctx, name)
}

// GetXML mocks base method.
func (m *MockPoolManager) GetXML(ctx context.Context, name string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetXML", ctx, name)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetXML indicates an expected call of GetXML.
func (mr *MockPoolManagerMockRecorder) GetXML(ctx, name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetXML", reflect.TypeOf((*MockPoolManager)(nil).GetXML), ctx, name)
}

// IsActive mocks base method.
func (m *MockPoolManager) IsActive(ctx context.Context, name string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsActive", ctx, name)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsActive indicates an expected call of IsActive.
func (mr *MockPoolManagerMockRecorder) IsActive(ctx, name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsActive", reflect.TypeOf((*MockPoolManager)(nil).IsActive), ctx, name)
}

// List mocks base method.
func (m *MockPoolManager) List(ctx context.Context) ([]*storage.StoragePoolInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", ctx)
	ret0, _ := ret[0].([]*storage.StoragePoolInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockPoolManagerMockRecorder) List(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockPoolManager)(nil).List), ctx)
}

// Refresh mocks base method.
func (m *MockPoolManager) Refresh(ctx context.Context, name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Refresh", ctx, name)
	ret0, _ := ret[0].(error)
	return ret0
}

// Refresh indicates an expected call of Refresh.
func (mr *MockPoolManagerMockRecorder) Refresh(ctx, name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Refresh", reflect.TypeOf((*MockPoolManager)(nil).Refresh), ctx, name)
}

// SetAutostart mocks base method.
func (m *MockPoolManager) SetAutostart(ctx context.Context, name string, autostart bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetAutostart", ctx, name, autostart)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetAutostart indicates an expected call of SetAutostart.
func (mr *MockPoolManagerMockRecorder) SetAutostart(ctx, name, autostart any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetAutostart", reflect.TypeOf((*MockPoolManager)(nil).SetAutostart), ctx, name, autostart)
}

// Start mocks base method.
func (m *MockPoolManager) Start(ctx context.Context, name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Start", ctx, name)
	ret0, _ := ret[0].(error)
	return ret0
}

// Start indicates an expected call of Start.
func (mr *MockPoolManagerMockRecorder) Start(ctx, name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockPoolManager)(nil).Start), ctx, name)
}

// Stop mocks base method.
func (m *MockPoolManager) Stop(ctx context.Context, name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stop", ctx, name)
	ret0, _ := ret[0].(error)
	return ret0
}

// Stop indicates an expected call of Stop.
func (mr *MockPoolManagerMockRecorder) Stop(ctx, name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockPoolManager)(nil).Stop), ctx, name)
}

// MockVolumeManager is a mock of VolumeManager interface.
type MockVolumeManager struct {
	isgomock struct{}
	ctrl     *gomock.Controller
	recorder *MockVolumeManagerMockRecorder
}

// MockVolumeManagerMockRecorder is the mock recorder for MockVolumeManager.
type MockVolumeManagerMockRecorder struct {
	mock *MockVolumeManager
}

// NewMockVolumeManager creates a new mock instance.
func NewMockVolumeManager(ctrl *gomock.Controller) *MockVolumeManager {
	mock := &MockVolumeManager{ctrl: ctrl}
	mock.recorder = &MockVolumeManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockVolumeManager) EXPECT() *MockVolumeManagerMockRecorder {
	return m.recorder
}

// Clone mocks base method.
func (m *MockVolumeManager) Clone(ctx context.Context, poolName, sourceVolName, destVolName string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Clone", ctx, poolName, sourceVolName, destVolName)
	ret0, _ := ret[0].(error)
	return ret0
}

// Clone indicates an expected call of Clone.
func (mr *MockVolumeManagerMockRecorder) Clone(ctx, poolName, sourceVolName, destVolName any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Clone", reflect.TypeOf((*MockVolumeManager)(nil).Clone), ctx, poolName, sourceVolName, destVolName)
}

// Create mocks base method.
func (m *MockVolumeManager) Create(ctx context.Context, poolName, volName string, capacityBytes uint64, format string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, poolName, volName, capacityBytes, format)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create.
func (mr *MockVolumeManagerMockRecorder) Create(ctx, poolName, volName, capacityBytes, format any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockVolumeManager)(nil).Create), ctx, poolName, volName, capacityBytes, format)
}

// CreateFromImage mocks base method.
func (m *MockVolumeManager) CreateFromImage(ctx context.Context, poolName, volName, imagePath, format string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateFromImage", ctx, poolName, volName, imagePath, format)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateFromImage indicates an expected call of CreateFromImage.
func (mr *MockVolumeManagerMockRecorder) CreateFromImage(ctx, poolName, volName, imagePath, format any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateFromImage", reflect.TypeOf((*MockVolumeManager)(nil).CreateFromImage), ctx, poolName, volName, imagePath, format)
}

// Delete mocks base method.
func (m *MockVolumeManager) Delete(ctx context.Context, poolName, volName string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, poolName, volName)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockVolumeManagerMockRecorder) Delete(ctx, poolName, volName any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockVolumeManager)(nil).Delete), ctx, poolName, volName)
}

// Download mocks base method.
func (m *MockVolumeManager) Download(ctx context.Context, poolName, volName string, writer io.Writer) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Download", ctx, poolName, volName, writer)
	ret0, _ := ret[0].(error)
	return ret0
}

// Download indicates an expected call of Download.
func (mr *MockVolumeManagerMockRecorder) Download(ctx, poolName, volName, writer any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Download", reflect.TypeOf((*MockVolumeManager)(nil).Download), ctx, poolName, volName, writer)
}

// GetInfo mocks base method.
func (m *MockVolumeManager) GetInfo(ctx context.Context, poolName, volName string) (*storage.StorageVolumeInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetInfo", ctx, poolName, volName)
	ret0, _ := ret[0].(*storage.StorageVolumeInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetInfo indicates an expected call of GetInfo.
func (mr *MockVolumeManagerMockRecorder) GetInfo(ctx, poolName, volName any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInfo", reflect.TypeOf((*MockVolumeManager)(nil).GetInfo), ctx, poolName, volName)
}

// GetPath mocks base method.
func (m *MockVolumeManager) GetPath(ctx context.Context, poolName, volName string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPath", ctx, poolName, volName)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPath indicates an expected call of GetPath.
func (mr *MockVolumeManagerMockRecorder) GetPath(ctx, poolName, volName any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPath", reflect.TypeOf((*MockVolumeManager)(nil).GetPath), ctx, poolName, volName)
}

// GetXML mocks base method.
func (m *MockVolumeManager) GetXML(ctx context.Context, poolName, volName string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetXML", ctx, poolName, volName)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetXML indicates an expected call of GetXML.
func (mr *MockVolumeManagerMockRecorder) GetXML(ctx, poolName, volName any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetXML", reflect.TypeOf((*MockVolumeManager)(nil).GetXML), ctx, poolName, volName)
}

// List mocks base method.
func (m *MockVolumeManager) List(ctx context.Context, poolName string) ([]*storage.StorageVolumeInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", ctx, poolName)
	ret0, _ := ret[0].([]*storage.StorageVolumeInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockVolumeManagerMockRecorder) List(ctx, poolName any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockVolumeManager)(nil).List), ctx, poolName)
}

// Resize mocks base method.
func (m *MockVolumeManager) Resize(ctx context.Context, poolName, volName string, capacityBytes uint64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Resize", ctx, poolName, volName, capacityBytes)
	ret0, _ := ret[0].(error)
	return ret0
}

// Resize indicates an expected call of Resize.
func (mr *MockVolumeManagerMockRecorder) Resize(ctx, poolName, volName, capacityBytes any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Resize", reflect.TypeOf((*MockVolumeManager)(nil).Resize), ctx, poolName, volName, capacityBytes)
}

// Upload mocks base method.
func (m *MockVolumeManager) Upload(ctx context.Context, poolName, volName string, reader io.Reader) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Upload", ctx, poolName, volName, reader)
	ret0, _ := ret[0].(error)
	return ret0
}

// Upload indicates an expected call of Upload.
func (mr *MockVolumeManagerMockRecorder) Upload(ctx, poolName, volName, reader any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Upload", reflect.TypeOf((*MockVolumeManager)(nil).Upload), ctx, poolName, volName, reader)
}

// Wipe mocks base method.
func (m *MockVolumeManager) Wipe(ctx context.Context, poolName, volName string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Wipe", ctx, poolName, volName)
	ret0, _ := ret[0].(error)
	return ret0
}

// Wipe indicates an expected call of Wipe.
func (mr *MockVolumeManagerMockRecorder) Wipe(ctx, poolName, volName any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Wipe", reflect.TypeOf((*MockVolumeManager)(nil).Wipe), ctx, poolName, volName)
}

// MockXMLBuilder is a mock of XMLBuilder interface.
type MockXMLBuilder struct {
	isgomock struct{}
	ctrl     *gomock.Controller
	recorder *MockXMLBuilderMockRecorder
}

// MockXMLBuilderMockRecorder is the mock recorder for MockXMLBuilder.
type MockXMLBuilderMockRecorder struct {
	mock *MockXMLBuilder
}

// NewMockXMLBuilder creates a new mock instance.
func NewMockXMLBuilder(ctrl *gomock.Controller) *MockXMLBuilder {
	mock := &MockXMLBuilder{ctrl: ctrl}
	mock.recorder = &MockXMLBuilderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockXMLBuilder) EXPECT() *MockXMLBuilderMockRecorder {
	return m.recorder
}

// BuildStoragePoolXML mocks base method.
func (m *MockXMLBuilder) BuildStoragePoolXML(name, path string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BuildStoragePoolXML", name, path)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BuildStoragePoolXML indicates an expected call of BuildStoragePoolXML.
func (mr *MockXMLBuilderMockRecorder) BuildStoragePoolXML(name, path any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BuildStoragePoolXML", reflect.TypeOf((*MockXMLBuilder)(nil).BuildStoragePoolXML), name, path)
}

// BuildStorageVolumeXML mocks base method.
func (m *MockXMLBuilder) BuildStorageVolumeXML(volName string, capacityBytes uint64, format string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BuildStorageVolumeXML", volName, capacityBytes, format)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BuildStorageVolumeXML indicates an expected call of BuildStorageVolumeXML.
func (mr *MockXMLBuilderMockRecorder) BuildStorageVolumeXML(volName, capacityBytes, format any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BuildStorageVolumeXML", reflect.TypeOf((*MockXMLBuilder)(nil).BuildStorageVolumeXML), volName, capacityBytes, format)
}
