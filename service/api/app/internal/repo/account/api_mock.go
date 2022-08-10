// Code generated by MockGen. DO NOT EDIT.
// Source: api.go

// Package account is a generated GoMock package.
package account

import (
	context "context"
	account "github.com/bnb-chain/zkbas/common/model/account"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockModel is a mock of Model interface
type MockModel struct {
	ctrl     *gomock.Controller
	recorder *MockModelMockRecorder
}

// MockModelMockRecorder is the mock recorder for MockModel
type MockModelMockRecorder struct {
	mock *MockModel
}

// NewMockModel creates a new mock instance
func NewMockModel(ctrl *gomock.Controller) *MockModel {
	mock := &MockModel{ctrl: ctrl}
	mock.recorder = &MockModelMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockModel) EXPECT() *MockModelMockRecorder {
	return m.recorder
}

// GetBasicAccountByAccountName mocks base method
func (m *MockModel) GetBasicAccountByAccountName(ctx context.Context, accountName string) (*account.Account, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBasicAccountByAccountName", ctx, accountName)
	ret0, _ := ret[0].(*account.Account)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBasicAccountByAccountName indicates an expected call of GetBasicAccountByAccountName
func (mr *MockModelMockRecorder) GetBasicAccountByAccountName(ctx, accountName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBasicAccountByAccountName", reflect.TypeOf((*MockModel)(nil).GetBasicAccountByAccountName), ctx, accountName)
}

// GetBasicAccountByAccountPk mocks base method
func (m *MockModel) GetBasicAccountByAccountPk(ctx context.Context, accountPK string) (*account.Account, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBasicAccountByAccountPk", ctx, accountPK)
	ret0, _ := ret[0].(*account.Account)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBasicAccountByAccountPk indicates an expected call of GetBasicAccountByAccountPk
func (mr *MockModelMockRecorder) GetBasicAccountByAccountPk(ctx, accountPK interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBasicAccountByAccountPk", reflect.TypeOf((*MockModel)(nil).GetBasicAccountByAccountPk), ctx, accountPK)
}

// GetAccountByAccountIndex mocks base method
func (m *MockModel) GetAccountByAccountIndex(accountIndex int64) (*account.Account, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAccountByAccountIndex", accountIndex)
	ret0, _ := ret[0].(*account.Account)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAccountByAccountIndex indicates an expected call of GetAccountByAccountIndex
func (mr *MockModelMockRecorder) GetAccountByAccountIndex(accountIndex interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAccountByAccountIndex", reflect.TypeOf((*MockModel)(nil).GetAccountByAccountIndex), accountIndex)
}

// GetAccountByPk mocks base method
func (m *MockModel) GetAccountByPk(pk string) (*account.Account, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAccountByPk", pk)
	ret0, _ := ret[0].(*account.Account)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAccountByPk indicates an expected call of GetAccountByPk
func (mr *MockModelMockRecorder) GetAccountByPk(pk interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAccountByPk", reflect.TypeOf((*MockModel)(nil).GetAccountByPk), pk)
}

// GetAccountByAccountName mocks base method
func (m *MockModel) GetAccountByAccountName(ctx context.Context, accountName string) (*account.Account, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAccountByAccountName", ctx, accountName)
	ret0, _ := ret[0].(*account.Account)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAccountByAccountName indicates an expected call of GetAccountByAccountName
func (mr *MockModelMockRecorder) GetAccountByAccountName(ctx, accountName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAccountByAccountName", reflect.TypeOf((*MockModel)(nil).GetAccountByAccountName), ctx, accountName)
}

// GetAccountsList mocks base method
func (m *MockModel) GetAccountsList(limit int, offset int64) ([]*account.Account, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAccountsList", limit, offset)
	ret0, _ := ret[0].([]*account.Account)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAccountsList indicates an expected call of GetAccountsList
func (mr *MockModelMockRecorder) GetAccountsList(limit, offset interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAccountsList", reflect.TypeOf((*MockModel)(nil).GetAccountsList), limit, offset)
}

// GetAccountsTotalCount mocks base method
func (m *MockModel) GetAccountsTotalCount() (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAccountsTotalCount")
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAccountsTotalCount indicates an expected call of GetAccountsTotalCount
func (mr *MockModelMockRecorder) GetAccountsTotalCount() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAccountsTotalCount", reflect.TypeOf((*MockModel)(nil).GetAccountsTotalCount))
}