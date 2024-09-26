// Code generated by MockGen. DO NOT EDIT.
// Source: webook/internal/repository/article/article_reader.go
//
// Generated by this command:
//
//	mockgen -source=webook/internal/repository/article/article_reader.go -package=artrepomocks -destination=webook/internal/repository/article/mocks/article_reader.mock.go
//

// Package artrepomocks is a generated GoMock package.
package artrepomocks

import (
	domain "awesomeProject/webook/internal/domain"
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockArticleReaderRepository is a mock of ArticleReaderRepository interface.
type MockArticleReaderRepository struct {
	ctrl     *gomock.Controller
	recorder *MockArticleReaderRepositoryMockRecorder
}

// MockArticleReaderRepositoryMockRecorder is the mock recorder for MockArticleReaderRepository.
type MockArticleReaderRepositoryMockRecorder struct {
	mock *MockArticleReaderRepository
}

// NewMockArticleReaderRepository creates a new mock instance.
func NewMockArticleReaderRepository(ctrl *gomock.Controller) *MockArticleReaderRepository {
	mock := &MockArticleReaderRepository{ctrl: ctrl}
	mock.recorder = &MockArticleReaderRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockArticleReaderRepository) EXPECT() *MockArticleReaderRepositoryMockRecorder {
	return m.recorder
}

// Save mocks base method.
func (m *MockArticleReaderRepository) Save(ctx context.Context, art domain.Article) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Save", ctx, art)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Save indicates an expected call of Save.
func (mr *MockArticleReaderRepositoryMockRecorder) Save(ctx, art any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Save", reflect.TypeOf((*MockArticleReaderRepository)(nil).Save), ctx, art)
}
