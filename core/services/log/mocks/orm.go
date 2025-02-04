// Code generated by mockery v2.8.0. DO NOT EDIT.

package mocks

import (
	common "github.com/ethereum/go-ethereum/common"
	log "github.com/smartcontractkit/chainlink/core/services/log"
	mock "github.com/stretchr/testify/mock"

	postgres "github.com/smartcontractkit/chainlink/core/services/postgres"
)

// ORM is an autogenerated mock type for the ORM type
type ORM struct {
	mock.Mock
}

// CreateBroadcast provides a mock function with given fields: blockHash, blockNumber, logIndex, jobID, qopts
func (_m *ORM) CreateBroadcast(blockHash common.Hash, blockNumber uint64, logIndex uint, jobID int32, qopts ...postgres.QOpt) error {
	_va := make([]interface{}, len(qopts))
	for _i := range qopts {
		_va[_i] = qopts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, blockHash, blockNumber, logIndex, jobID)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(common.Hash, uint64, uint, int32, ...postgres.QOpt) error); ok {
		r0 = rf(blockHash, blockNumber, logIndex, jobID, qopts...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// FindBroadcasts provides a mock function with given fields: fromBlockNum, toBlockNum
func (_m *ORM) FindBroadcasts(fromBlockNum int64, toBlockNum int64) ([]log.LogBroadcast, error) {
	ret := _m.Called(fromBlockNum, toBlockNum)

	var r0 []log.LogBroadcast
	if rf, ok := ret.Get(0).(func(int64, int64) []log.LogBroadcast); ok {
		r0 = rf(fromBlockNum, toBlockNum)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]log.LogBroadcast)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int64, int64) error); ok {
		r1 = rf(fromBlockNum, toBlockNum)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPendingMinBlock provides a mock function with given fields: qopts
func (_m *ORM) GetPendingMinBlock(qopts ...postgres.QOpt) (*int64, error) {
	_va := make([]interface{}, len(qopts))
	for _i := range qopts {
		_va[_i] = qopts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *int64
	if rf, ok := ret.Get(0).(func(...postgres.QOpt) *int64); ok {
		r0 = rf(qopts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*int64)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(...postgres.QOpt) error); ok {
		r1 = rf(qopts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MarkBroadcastConsumed provides a mock function with given fields: blockHash, blockNumber, logIndex, jobID, qopts
func (_m *ORM) MarkBroadcastConsumed(blockHash common.Hash, blockNumber uint64, logIndex uint, jobID int32, qopts ...postgres.QOpt) error {
	_va := make([]interface{}, len(qopts))
	for _i := range qopts {
		_va[_i] = qopts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, blockHash, blockNumber, logIndex, jobID)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(common.Hash, uint64, uint, int32, ...postgres.QOpt) error); ok {
		r0 = rf(blockHash, blockNumber, logIndex, jobID, qopts...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Reinitialize provides a mock function with given fields: qopts
func (_m *ORM) Reinitialize(qopts ...postgres.QOpt) (*int64, error) {
	_va := make([]interface{}, len(qopts))
	for _i := range qopts {
		_va[_i] = qopts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *int64
	if rf, ok := ret.Get(0).(func(...postgres.QOpt) *int64); ok {
		r0 = rf(qopts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*int64)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(...postgres.QOpt) error); ok {
		r1 = rf(qopts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetPendingMinBlock provides a mock function with given fields: blockNum, qopts
func (_m *ORM) SetPendingMinBlock(blockNum *int64, qopts ...postgres.QOpt) error {
	_va := make([]interface{}, len(qopts))
	for _i := range qopts {
		_va[_i] = qopts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, blockNum)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(*int64, ...postgres.QOpt) error); ok {
		r0 = rf(blockNum, qopts...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// WasBroadcastConsumed provides a mock function with given fields: blockHash, logIndex, jobID, qopts
func (_m *ORM) WasBroadcastConsumed(blockHash common.Hash, logIndex uint, jobID int32, qopts ...postgres.QOpt) (bool, error) {
	_va := make([]interface{}, len(qopts))
	for _i := range qopts {
		_va[_i] = qopts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, blockHash, logIndex, jobID)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 bool
	if rf, ok := ret.Get(0).(func(common.Hash, uint, int32, ...postgres.QOpt) bool); ok {
		r0 = rf(blockHash, logIndex, jobID, qopts...)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(common.Hash, uint, int32, ...postgres.QOpt) error); ok {
		r1 = rf(blockHash, logIndex, jobID, qopts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
