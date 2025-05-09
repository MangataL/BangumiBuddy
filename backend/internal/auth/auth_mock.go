// Code generated by MockGen. DO NOT EDIT.
// Source: auth.go

// Package auth is a generated GoMock package.
package auth

import (
	context "context"
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
)

// MockCipher is a mock of Cipher interface.
type MockCipher struct {
	ctrl     *gomock.Controller
	recorder *MockCipherMockRecorder
}

// MockCipherMockRecorder is the mock recorder for MockCipher.
type MockCipherMockRecorder struct {
	mock *MockCipher
}

// NewMockCipher creates a new mock instance.
func NewMockCipher(ctrl *gomock.Controller) *MockCipher {
	mock := &MockCipher{ctrl: ctrl}
	mock.recorder = &MockCipherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCipher) EXPECT() *MockCipherMockRecorder {
	return m.recorder
}

// Check mocks base method.
func (m *MockCipher) Check(ctx context.Context, key, text, cipher string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Check", ctx, key, text, cipher)
	ret0, _ := ret[0].(error)
	return ret0
}

// Check indicates an expected call of Check.
func (mr *MockCipherMockRecorder) Check(ctx, key, text, cipher interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Check", reflect.TypeOf((*MockCipher)(nil).Check), ctx, key, text, cipher)
}

// Encrypt mocks base method.
func (m *MockCipher) Encrypt(ctx context.Context, key, text string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Encrypt", ctx, key, text)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Encrypt indicates an expected call of Encrypt.
func (mr *MockCipherMockRecorder) Encrypt(ctx, key, text interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Encrypt", reflect.TypeOf((*MockCipher)(nil).Encrypt), ctx, key, text)
}

// GenerateKey mocks base method.
func (m *MockCipher) GenerateKey(ctx context.Context) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateKey", ctx)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GenerateKey indicates an expected call of GenerateKey.
func (mr *MockCipherMockRecorder) GenerateKey(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateKey", reflect.TypeOf((*MockCipher)(nil).GenerateKey), ctx)
}

// MockTokenOperator is a mock of TokenOperator interface.
type MockTokenOperator struct {
	ctrl     *gomock.Controller
	recorder *MockTokenOperatorMockRecorder
}

// MockTokenOperatorMockRecorder is the mock recorder for MockTokenOperator.
type MockTokenOperatorMockRecorder struct {
	mock *MockTokenOperator
}

// NewMockTokenOperator creates a new mock instance.
func NewMockTokenOperator(ctrl *gomock.Controller) *MockTokenOperator {
	mock := &MockTokenOperator{ctrl: ctrl}
	mock.recorder = &MockTokenOperatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTokenOperator) EXPECT() *MockTokenOperatorMockRecorder {
	return m.recorder
}

// Check mocks base method.
func (m *MockTokenOperator) Check(ctx context.Context, tokenType, key, token string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Check", ctx, tokenType, key, token)
	ret0, _ := ret[0].(error)
	return ret0
}

// Check indicates an expected call of Check.
func (mr *MockTokenOperatorMockRecorder) Check(ctx, tokenType, key, token interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Check", reflect.TypeOf((*MockTokenOperator)(nil).Check), ctx, tokenType, key, token)
}

// Generate mocks base method.
func (m *MockTokenOperator) Generate(ctx context.Context, tokenType, key string, expireAt time.Time) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Generate", ctx, tokenType, key, expireAt)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Generate indicates an expected call of Generate.
func (mr *MockTokenOperatorMockRecorder) Generate(ctx, tokenType, key, expireAt interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Generate", reflect.TypeOf((*MockTokenOperator)(nil).Generate), ctx, tokenType, key, expireAt)
}

// MockConfig is a mock of Config interface.
type MockConfig struct {
	ctrl     *gomock.Controller
	recorder *MockConfigMockRecorder
}

// MockConfigMockRecorder is the mock recorder for MockConfig.
type MockConfigMockRecorder struct {
	mock *MockConfig
}

// NewMockConfig creates a new mock instance.
func NewMockConfig(ctrl *gomock.Controller) *MockConfig {
	mock := &MockConfig{ctrl: ctrl}
	mock.recorder = &MockConfigMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockConfig) EXPECT() *MockConfigMockRecorder {
	return m.recorder
}

// GetPassword mocks base method.
func (m *MockConfig) GetPassword() (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPassword")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPassword indicates an expected call of GetPassword.
func (mr *MockConfigMockRecorder) GetPassword() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPassword", reflect.TypeOf((*MockConfig)(nil).GetPassword))
}

// GetUserToken mocks base method.
func (m *MockConfig) GetUserToken() (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserToken")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserToken indicates an expected call of GetUserToken.
func (mr *MockConfigMockRecorder) GetUserToken() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserToken", reflect.TypeOf((*MockConfig)(nil).GetUserToken))
}

// GetUsername mocks base method.
func (m *MockConfig) GetUsername() (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUsername")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUsername indicates an expected call of GetUsername.
func (mr *MockConfigMockRecorder) GetUsername() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUsername", reflect.TypeOf((*MockConfig)(nil).GetUsername))
}

// SetPassword mocks base method.
func (m *MockConfig) SetPassword(password string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetPassword", password)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetPassword indicates an expected call of SetPassword.
func (mr *MockConfigMockRecorder) SetPassword(password interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetPassword", reflect.TypeOf((*MockConfig)(nil).SetPassword), password)
}

// SetUserToken mocks base method.
func (m *MockConfig) SetUserToken(token string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetUserToken", token)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetUserToken indicates an expected call of SetUserToken.
func (mr *MockConfigMockRecorder) SetUserToken(token interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetUserToken", reflect.TypeOf((*MockConfig)(nil).SetUserToken), token)
}

// SetUsername mocks base method.
func (m *MockConfig) SetUsername(username string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetUsername", username)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetUsername indicates an expected call of SetUsername.
func (mr *MockConfigMockRecorder) SetUsername(username interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetUsername", reflect.TypeOf((*MockConfig)(nil).SetUsername), username)
}
