// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/StasMerzlyakov/gophermart/internal/gophermart/app (interfaces: OrderStorage,AcrualSystem)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	domain "github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
	gomock "github.com/golang/mock/gomock"
)

// MockOrderStorage is a mock of OrderStorage interface.
type MockOrderStorage struct {
	ctrl     *gomock.Controller
	recorder *MockOrderStorageMockRecorder
}

// MockOrderStorageMockRecorder is the mock recorder for MockOrderStorage.
type MockOrderStorageMockRecorder struct {
	mock *MockOrderStorage
}

// NewMockOrderStorage creates a new mock instance.
func NewMockOrderStorage(ctrl *gomock.Controller) *MockOrderStorage {
	mock := &MockOrderStorage{ctrl: ctrl}
	mock.recorder = &MockOrderStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockOrderStorage) EXPECT() *MockOrderStorageMockRecorder {
	return m.recorder
}

// GetByStatus mocks base method.
func (m *MockOrderStorage) GetByStatus(arg0 context.Context, arg1 domain.OrderStatus) ([]domain.OrderData, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByStatus", arg0, arg1)
	ret0, _ := ret[0].([]domain.OrderData)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByStatus indicates an expected call of GetByStatus.
func (mr *MockOrderStorageMockRecorder) GetByStatus(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByStatus", reflect.TypeOf((*MockOrderStorage)(nil).GetByStatus), arg0, arg1)
}

// Orders mocks base method.
func (m *MockOrderStorage) Orders(arg0 context.Context, arg1 int) ([]domain.OrderData, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Orders", arg0, arg1)
	ret0, _ := ret[0].([]domain.OrderData)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Orders indicates an expected call of Orders.
func (mr *MockOrderStorageMockRecorder) Orders(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Orders", reflect.TypeOf((*MockOrderStorage)(nil).Orders), arg0, arg1)
}

// UpdateOrders mocks base method.
func (m *MockOrderStorage) UpdateOrders(arg0 context.Context, arg1 []domain.OrderData) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateOrders", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateOrders indicates an expected call of UpdateOrders.
func (mr *MockOrderStorageMockRecorder) UpdateOrders(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateOrders", reflect.TypeOf((*MockOrderStorage)(nil).UpdateOrders), arg0, arg1)
}

// Upload mocks base method.
func (m *MockOrderStorage) Upload(arg0 context.Context, arg1 *domain.OrderData) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Upload", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Upload indicates an expected call of Upload.
func (mr *MockOrderStorageMockRecorder) Upload(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Upload", reflect.TypeOf((*MockOrderStorage)(nil).Upload), arg0, arg1)
}

// MockAcrualSystem is a mock of AcrualSystem interface.
type MockAcrualSystem struct {
	ctrl     *gomock.Controller
	recorder *MockAcrualSystemMockRecorder
}

// MockAcrualSystemMockRecorder is the mock recorder for MockAcrualSystem.
type MockAcrualSystemMockRecorder struct {
	mock *MockAcrualSystem
}

// NewMockAcrualSystem creates a new mock instance.
func NewMockAcrualSystem(ctrl *gomock.Controller) *MockAcrualSystem {
	mock := &MockAcrualSystem{ctrl: ctrl}
	mock.recorder = &MockAcrualSystemMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAcrualSystem) EXPECT() *MockAcrualSystemMockRecorder {
	return m.recorder
}

// GetStatus mocks base method.
func (m *MockAcrualSystem) GetStatus(arg0 context.Context, arg1 domain.OrderNumber) (*domain.AccrualData, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStatus", arg0, arg1)
	ret0, _ := ret[0].(*domain.AccrualData)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetStatus indicates an expected call of GetStatus.
func (mr *MockAcrualSystemMockRecorder) GetStatus(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStatus", reflect.TypeOf((*MockAcrualSystem)(nil).GetStatus), arg0, arg1)
}
