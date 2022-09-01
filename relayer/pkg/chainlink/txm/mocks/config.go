// Code generated by mockery v2.13.0-beta.1. DO NOT EDIT.

package mocks

import (
	time "time"

	mock "github.com/stretchr/testify/mock"
)

// Config is an autogenerated mock type for the Config type
type Config struct {
	mock.Mock
}

// TxConfirmFrequency provides a mock function with given fields:
func (_m *Config) TxConfirmFrequency() time.Duration {
	ret := _m.Called()

	var r0 time.Duration
	if rf, ok := ret.Get(0).(func() time.Duration); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(time.Duration)
	}

	return r0
}

// TxRetryFrequency provides a mock function with given fields:
func (_m *Config) TxRetryFrequency() time.Duration {
	ret := _m.Called()

	var r0 time.Duration
	if rf, ok := ret.Get(0).(func() time.Duration); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(time.Duration)
	}

	return r0
}

// TxTimeout provides a mock function with given fields:
func (_m *Config) TxTimeout() time.Duration {
	ret := _m.Called()

	var r0 time.Duration
	if rf, ok := ret.Get(0).(func() time.Duration); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(time.Duration)
	}

	return r0
}

type NewConfigT interface {
	mock.TestingT
	Cleanup(func())
}

// NewConfig creates a new instance of Config. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewConfig(t NewConfigT) *Config {
	mock := &Config{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
