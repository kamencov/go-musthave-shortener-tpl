// Code generated by MockGen. DO NOT EDIT.
// Source: ./contract.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	models "github.com/kamencov/go-musthave-shortener-tpl/internal/models"
)

// MockStorage is a mock of Storage interface.
type MockStorage struct {
	ctrl     *gomock.Controller
	recorder *MockStorageMockRecorder
}

// MockStorageMockRecorder is the mock recorder for MockStorage.
type MockStorageMockRecorder struct {
	mock *MockStorage
}

// NewMockStorage creates a new mock instance.
func NewMockStorage(ctrl *gomock.Controller) *MockStorage {
	mock := &MockStorage{ctrl: ctrl}
	mock.recorder = &MockStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStorage) EXPECT() *MockStorageMockRecorder {
	return m.recorder
}

// CheckURL mocks base method.
func (m *MockStorage) CheckURL(arg0 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckURL", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CheckURL indicates an expected call of CheckURL.
func (mr *MockStorageMockRecorder) CheckURL(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckURL", reflect.TypeOf((*MockStorage)(nil).CheckURL), arg0)
}

// Close mocks base method.
func (m *MockStorage) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockStorageMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockStorage)(nil).Close))
}

// DeletedURLs mocks base method.
func (m *MockStorage) DeletedURLs(urls []string, userID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeletedURLs", urls, userID)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeletedURLs indicates an expected call of DeletedURLs.
func (mr *MockStorageMockRecorder) DeletedURLs(urls, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeletedURLs", reflect.TypeOf((*MockStorage)(nil).DeletedURLs), urls, userID)
}

// GetAllURL mocks base method.
func (m *MockStorage) GetAllURL(userID, baseURL string) ([]*models.UserURLs, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllURL", userID, baseURL)
	ret0, _ := ret[0].([]*models.UserURLs)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllURL indicates an expected call of GetAllURL.
func (mr *MockStorageMockRecorder) GetAllURL(userID, baseURL interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllURL", reflect.TypeOf((*MockStorage)(nil).GetAllURL), userID, baseURL)
}

// GetURL mocks base method.
func (m *MockStorage) GetURL(arg0 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetURL", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetURL indicates an expected call of GetURL.
func (mr *MockStorageMockRecorder) GetURL(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetURL", reflect.TypeOf((*MockStorage)(nil).GetURL), arg0)
}

// Ping mocks base method.
func (m *MockStorage) Ping() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ping")
	ret0, _ := ret[0].(error)
	return ret0
}

// Ping indicates an expected call of Ping.
func (mr *MockStorageMockRecorder) Ping() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ping", reflect.TypeOf((*MockStorage)(nil).Ping))
}

// SaveSliceOfDB mocks base method.
func (m *MockStorage) SaveSlice(urls []models.MultipleURL, baseURL, userID string) ([]models.ResultMultipleURL, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveSlice", urls, baseURL, userID)
	ret0, _ := ret[0].([]models.ResultMultipleURL)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SaveSliceOfDB indicates an expected call of SaveSliceOfDB.
func (mr *MockStorageMockRecorder) SaveSliceOfDB(urls, baseURL, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveSlice", reflect.TypeOf((*MockStorage)(nil).SaveSlice), urls, baseURL, userID)
}

// SaveURL mocks base method.
func (m *MockStorage) SaveURL(shortURL, originalURL, userID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveURL", shortURL, originalURL, userID)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveURL indicates an expected call of SaveURL.
func (mr *MockStorageMockRecorder) SaveURL(shortURL, originalURL, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveURL", reflect.TypeOf((*MockStorage)(nil).SaveURL), shortURL, originalURL, userID)
}