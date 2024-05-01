// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/StasMerzlyakov/gophermart/internal/gophermart/adapter/in/http/handler/order (interfaces: CreateOrderApp)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	domain "github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
	gomock "github.com/golang/mock/gomock"
)

// MockCreateOrderApp is a mock of CreateOrderApp interface.
type MockCreateOrderApp struct {
	ctrl     *gomock.Controller
	recorder *MockCreateOrderAppMockRecorder
}

// MockCreateOrderAppMockRecorder is the mock recorder for MockCreateOrderApp.
type MockCreateOrderAppMockRecorder struct {
	mock *MockCreateOrderApp
}

// NewMockCreateOrderApp creates a new mock instance.
func NewMockCreateOrderApp(ctrl *gomock.Controller) *MockCreateOrderApp {
	mock := &MockCreateOrderApp{ctrl: ctrl}
	mock.recorder = &MockCreateOrderAppMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCreateOrderApp) EXPECT() *MockCreateOrderAppMockRecorder {
	return m.recorder
}

// New mocks base method.
func (m *MockCreateOrderApp) New(arg0 context.Context, arg1 domain.OrderNumber) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "New", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// New indicates an expected call of New.
func (mr *MockCreateOrderAppMockRecorder) New(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "New", reflect.TypeOf((*MockCreateOrderApp)(nil).New), arg0, arg1)
}
