// Code generated by mockery. DO NOT EDIT.

package api

import (
	pagefilter "github.com/jacobbrewer1/pagefilter"
	models "github.com/jacobbrewer1/puppet-reporter/pkg/models"
	mock "github.com/stretchr/testify/mock"
)

// MockRepository is an autogenerated mock type for the Repository type
type MockRepository struct {
	mock.Mock
}

// GetLogsByReportID provides a mock function with given fields: reportID
func (_m *MockRepository) GetLogsByReportID(reportID int) ([]*models.LogMessage, error) {
	ret := _m.Called(reportID)

	if len(ret) == 0 {
		panic("no return value specified for GetLogsByReportID")
	}

	var r0 []*models.LogMessage
	var r1 error
	if rf, ok := ret.Get(0).(func(int) ([]*models.LogMessage, error)); ok {
		return rf(reportID)
	}
	if rf, ok := ret.Get(0).(func(int) []*models.LogMessage); ok {
		r0 = rf(reportID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*models.LogMessage)
		}
	}

	if rf, ok := ret.Get(1).(func(int) error); ok {
		r1 = rf(reportID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetReportByHash provides a mock function with given fields: hash
func (_m *MockRepository) GetReportByHash(hash string) (*models.Report, error) {
	ret := _m.Called(hash)

	if len(ret) == 0 {
		panic("no return value specified for GetReportByHash")
	}

	var r0 *models.Report
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (*models.Report, error)); ok {
		return rf(hash)
	}
	if rf, ok := ret.Get(0).(func(string) *models.Report); ok {
		r0 = rf(hash)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Report)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(hash)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetReports provides a mock function with given fields: paginationDetails, filters
func (_m *MockRepository) GetReports(paginationDetails *pagefilter.PaginatorDetails, filters *GetReportsFilters) (*pagefilter.PaginatedResponse[models.Report], error) {
	ret := _m.Called(paginationDetails, filters)

	if len(ret) == 0 {
		panic("no return value specified for GetReports")
	}

	var r0 *pagefilter.PaginatedResponse[models.Report]
	var r1 error
	if rf, ok := ret.Get(0).(func(*pagefilter.PaginatorDetails, *GetReportsFilters) (*pagefilter.PaginatedResponse[models.Report], error)); ok {
		return rf(paginationDetails, filters)
	}
	if rf, ok := ret.Get(0).(func(*pagefilter.PaginatorDetails, *GetReportsFilters) *pagefilter.PaginatedResponse[models.Report]); ok {
		r0 = rf(paginationDetails, filters)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*pagefilter.PaginatedResponse[models.Report])
		}
	}

	if rf, ok := ret.Get(1).(func(*pagefilter.PaginatorDetails, *GetReportsFilters) error); ok {
		r1 = rf(paginationDetails, filters)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetResourcesByReportID provides a mock function with given fields: reportID
func (_m *MockRepository) GetResourcesByReportID(reportID int) ([]*models.Resource, error) {
	ret := _m.Called(reportID)

	if len(ret) == 0 {
		panic("no return value specified for GetResourcesByReportID")
	}

	var r0 []*models.Resource
	var r1 error
	if rf, ok := ret.Get(0).(func(int) ([]*models.Resource, error)); ok {
		return rf(reportID)
	}
	if rf, ok := ret.Get(0).(func(int) []*models.Resource); ok {
		r0 = rf(reportID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*models.Resource)
		}
	}

	if rf, ok := ret.Get(1).(func(int) error); ok {
		r1 = rf(reportID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SaveLogs provides a mock function with given fields: logs
func (_m *MockRepository) SaveLogs(logs []*models.LogMessage) error {
	ret := _m.Called(logs)

	if len(ret) == 0 {
		panic("no return value specified for SaveLogs")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func([]*models.LogMessage) error); ok {
		r0 = rf(logs)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SaveReport provides a mock function with given fields: report
func (_m *MockRepository) SaveReport(report *models.Report) error {
	ret := _m.Called(report)

	if len(ret) == 0 {
		panic("no return value specified for SaveReport")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*models.Report) error); ok {
		r0 = rf(report)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SaveResources provides a mock function with given fields: resources
func (_m *MockRepository) SaveResources(resources []*models.Resource) error {
	ret := _m.Called(resources)

	if len(ret) == 0 {
		panic("no return value specified for SaveResources")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func([]*models.Resource) error); ok {
		r0 = rf(resources)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewMockRepository creates a new instance of MockRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockRepository {
	mock := &MockRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
