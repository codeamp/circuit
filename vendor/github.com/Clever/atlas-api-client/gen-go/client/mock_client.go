// Code generated by MockGen. DO NOT EDIT.
// Source: interface.go

// Package client is a generated GoMock package.
package client

import (
	context "context"
	models "github.com/Clever/atlas-api-client/gen-go/models"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockClient is a mock of Client interface
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// GetClusters mocks base method
func (m *MockClient) GetClusters(ctx context.Context, groupID string) (*models.GetClustersResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetClusters", ctx, groupID)
	ret0, _ := ret[0].(*models.GetClustersResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetClusters indicates an expected call of GetClusters
func (mr *MockClientMockRecorder) GetClusters(ctx, groupID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetClusters", reflect.TypeOf((*MockClient)(nil).GetClusters), ctx, groupID)
}

// CreateCluster mocks base method
func (m *MockClient) CreateCluster(ctx context.Context, i *models.CreateClusterInput) (*models.Cluster, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateCluster", ctx, i)
	ret0, _ := ret[0].(*models.Cluster)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateCluster indicates an expected call of CreateCluster
func (mr *MockClientMockRecorder) CreateCluster(ctx, i interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateCluster", reflect.TypeOf((*MockClient)(nil).CreateCluster), ctx, i)
}

// DeleteCluster mocks base method
func (m *MockClient) DeleteCluster(ctx context.Context, i *models.DeleteClusterInput) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteCluster", ctx, i)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteCluster indicates an expected call of DeleteCluster
func (mr *MockClientMockRecorder) DeleteCluster(ctx, i interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteCluster", reflect.TypeOf((*MockClient)(nil).DeleteCluster), ctx, i)
}

// GetCluster mocks base method
func (m *MockClient) GetCluster(ctx context.Context, i *models.GetClusterInput) (*models.Cluster, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCluster", ctx, i)
	ret0, _ := ret[0].(*models.Cluster)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCluster indicates an expected call of GetCluster
func (mr *MockClientMockRecorder) GetCluster(ctx, i interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCluster", reflect.TypeOf((*MockClient)(nil).GetCluster), ctx, i)
}

// UpdateCluster mocks base method
func (m *MockClient) UpdateCluster(ctx context.Context, i *models.UpdateClusterInput) (*models.Cluster, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateCluster", ctx, i)
	ret0, _ := ret[0].(*models.Cluster)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateCluster indicates an expected call of UpdateCluster
func (mr *MockClientMockRecorder) UpdateCluster(ctx, i interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateCluster", reflect.TypeOf((*MockClient)(nil).UpdateCluster), ctx, i)
}

// GetContainers mocks base method
func (m *MockClient) GetContainers(ctx context.Context, groupID string) (*models.GetContainersResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetContainers", ctx, groupID)
	ret0, _ := ret[0].(*models.GetContainersResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetContainers indicates an expected call of GetContainers
func (mr *MockClientMockRecorder) GetContainers(ctx, groupID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetContainers", reflect.TypeOf((*MockClient)(nil).GetContainers), ctx, groupID)
}

// CreateContainer mocks base method
func (m *MockClient) CreateContainer(ctx context.Context, i *models.CreateContainerInput) (*models.Container, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateContainer", ctx, i)
	ret0, _ := ret[0].(*models.Container)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateContainer indicates an expected call of CreateContainer
func (mr *MockClientMockRecorder) CreateContainer(ctx, i interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateContainer", reflect.TypeOf((*MockClient)(nil).CreateContainer), ctx, i)
}

// GetContainer mocks base method
func (m *MockClient) GetContainer(ctx context.Context, i *models.GetContainerInput) (*models.Container, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetContainer", ctx, i)
	ret0, _ := ret[0].(*models.Container)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetContainer indicates an expected call of GetContainer
func (mr *MockClientMockRecorder) GetContainer(ctx, i interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetContainer", reflect.TypeOf((*MockClient)(nil).GetContainer), ctx, i)
}

// UpdateContainer mocks base method
func (m *MockClient) UpdateContainer(ctx context.Context, i *models.UpdateContainerInput) (*models.Container, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateContainer", ctx, i)
	ret0, _ := ret[0].(*models.Container)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateContainer indicates an expected call of UpdateContainer
func (mr *MockClientMockRecorder) UpdateContainer(ctx, i interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateContainer", reflect.TypeOf((*MockClient)(nil).UpdateContainer), ctx, i)
}

// GetDatabaseUsers mocks base method
func (m *MockClient) GetDatabaseUsers(ctx context.Context, groupID string) (*models.GetDatabaseUsersResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDatabaseUsers", ctx, groupID)
	ret0, _ := ret[0].(*models.GetDatabaseUsersResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDatabaseUsers indicates an expected call of GetDatabaseUsers
func (mr *MockClientMockRecorder) GetDatabaseUsers(ctx, groupID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDatabaseUsers", reflect.TypeOf((*MockClient)(nil).GetDatabaseUsers), ctx, groupID)
}

// CreateDatabaseUser mocks base method
func (m *MockClient) CreateDatabaseUser(ctx context.Context, i *models.CreateDatabaseUserInput) (*models.DatabaseUser, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateDatabaseUser", ctx, i)
	ret0, _ := ret[0].(*models.DatabaseUser)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateDatabaseUser indicates an expected call of CreateDatabaseUser
func (mr *MockClientMockRecorder) CreateDatabaseUser(ctx, i interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateDatabaseUser", reflect.TypeOf((*MockClient)(nil).CreateDatabaseUser), ctx, i)
}

// DeleteDatabaseUser mocks base method
func (m *MockClient) DeleteDatabaseUser(ctx context.Context, i *models.DeleteDatabaseUserInput) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteDatabaseUser", ctx, i)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteDatabaseUser indicates an expected call of DeleteDatabaseUser
func (mr *MockClientMockRecorder) DeleteDatabaseUser(ctx, i interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteDatabaseUser", reflect.TypeOf((*MockClient)(nil).DeleteDatabaseUser), ctx, i)
}

// GetDatabaseUser mocks base method
func (m *MockClient) GetDatabaseUser(ctx context.Context, i *models.GetDatabaseUserInput) (*models.DatabaseUser, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDatabaseUser", ctx, i)
	ret0, _ := ret[0].(*models.DatabaseUser)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDatabaseUser indicates an expected call of GetDatabaseUser
func (mr *MockClientMockRecorder) GetDatabaseUser(ctx, i interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDatabaseUser", reflect.TypeOf((*MockClient)(nil).GetDatabaseUser), ctx, i)
}

// UpdateDatabaseUser mocks base method
func (m *MockClient) UpdateDatabaseUser(ctx context.Context, i *models.UpdateDatabaseUserInput) (*models.DatabaseUser, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateDatabaseUser", ctx, i)
	ret0, _ := ret[0].(*models.DatabaseUser)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateDatabaseUser indicates an expected call of UpdateDatabaseUser
func (mr *MockClientMockRecorder) UpdateDatabaseUser(ctx, i interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateDatabaseUser", reflect.TypeOf((*MockClient)(nil).UpdateDatabaseUser), ctx, i)
}

// GetProcesses mocks base method
func (m *MockClient) GetProcesses(ctx context.Context, groupID string) (*models.GetProcessesResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProcesses", ctx, groupID)
	ret0, _ := ret[0].(*models.GetProcessesResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProcesses indicates an expected call of GetProcesses
func (mr *MockClientMockRecorder) GetProcesses(ctx, groupID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProcesses", reflect.TypeOf((*MockClient)(nil).GetProcesses), ctx, groupID)
}

// GetProcessDatabases mocks base method
func (m *MockClient) GetProcessDatabases(ctx context.Context, i *models.GetProcessDatabasesInput) (*models.GetProcessDatabasesResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProcessDatabases", ctx, i)
	ret0, _ := ret[0].(*models.GetProcessDatabasesResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProcessDatabases indicates an expected call of GetProcessDatabases
func (mr *MockClientMockRecorder) GetProcessDatabases(ctx, i interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProcessDatabases", reflect.TypeOf((*MockClient)(nil).GetProcessDatabases), ctx, i)
}

// GetProcessDatabaseMeasurements mocks base method
func (m *MockClient) GetProcessDatabaseMeasurements(ctx context.Context, i *models.GetProcessDatabaseMeasurementsInput) (*models.GetProcessDatabaseMeasurementsResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProcessDatabaseMeasurements", ctx, i)
	ret0, _ := ret[0].(*models.GetProcessDatabaseMeasurementsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProcessDatabaseMeasurements indicates an expected call of GetProcessDatabaseMeasurements
func (mr *MockClientMockRecorder) GetProcessDatabaseMeasurements(ctx, i interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProcessDatabaseMeasurements", reflect.TypeOf((*MockClient)(nil).GetProcessDatabaseMeasurements), ctx, i)
}

// GetProcessDisks mocks base method
func (m *MockClient) GetProcessDisks(ctx context.Context, i *models.GetProcessDisksInput) (*models.GetProcessDisksResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProcessDisks", ctx, i)
	ret0, _ := ret[0].(*models.GetProcessDisksResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProcessDisks indicates an expected call of GetProcessDisks
func (mr *MockClientMockRecorder) GetProcessDisks(ctx, i interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProcessDisks", reflect.TypeOf((*MockClient)(nil).GetProcessDisks), ctx, i)
}

// GetProcessDiskMeasurements mocks base method
func (m *MockClient) GetProcessDiskMeasurements(ctx context.Context, i *models.GetProcessDiskMeasurementsInput) (*models.GetProcessDiskMeasurementsResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProcessDiskMeasurements", ctx, i)
	ret0, _ := ret[0].(*models.GetProcessDiskMeasurementsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProcessDiskMeasurements indicates an expected call of GetProcessDiskMeasurements
func (mr *MockClientMockRecorder) GetProcessDiskMeasurements(ctx, i interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProcessDiskMeasurements", reflect.TypeOf((*MockClient)(nil).GetProcessDiskMeasurements), ctx, i)
}

// GetProcessMeasurements mocks base method
func (m *MockClient) GetProcessMeasurements(ctx context.Context, i *models.GetProcessMeasurementsInput) (*models.GetProcessMeasurementsResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProcessMeasurements", ctx, i)
	ret0, _ := ret[0].(*models.GetProcessMeasurementsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProcessMeasurements indicates an expected call of GetProcessMeasurements
func (mr *MockClientMockRecorder) GetProcessMeasurements(ctx, i interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProcessMeasurements", reflect.TypeOf((*MockClient)(nil).GetProcessMeasurements), ctx, i)
}
