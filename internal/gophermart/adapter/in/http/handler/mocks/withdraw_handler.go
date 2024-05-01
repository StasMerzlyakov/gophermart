// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/StasMerzlyakov/gophermart/internal/gophermart/adapter/in/http/handler/balance (interfaces: WithdrawApp)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	domain "github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
	gomock "github.com/golang/mock/gomock"
)

// MockWithdrawApp is a mock of WithdrawApp interface.
type MockWithdrawApp struct {
	ctrl     *gomock.Controller
	recorder *MockWithdrawAppMockRecorder
}

// MockWithdrawAppMockRecorder is the mock recorder for MockWithdrawApp.
type MockWithdrawAppMockRecorder struct {
	mock *MockWithdrawApp
}

// NewMockWithdrawApp creates a new mock instance.
func NewMockWithdrawApp(ctrl *gomock.Controller) *MockWithdrawApp {
	mock := &MockWithdrawApp{ctrl: ctrl}
	mock.recorder = &MockWithdrawAppMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWithdrawApp) EXPECT() *MockWithdrawAppMockRecorder {
	return m.recorder
}

// Withdraw mocks base method.
func (m *MockWithdrawApp) Withdraw(arg0 context.Context, arg1 *domain.WithdrawData) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Withdraw", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Withdraw indicates an expected call of Withdraw.
func (mr *MockWithdrawAppMockRecorder) Withdraw(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Withdraw", reflect.TypeOf((*MockWithdrawApp)(nil).Withdraw), arg0, arg1)
}
