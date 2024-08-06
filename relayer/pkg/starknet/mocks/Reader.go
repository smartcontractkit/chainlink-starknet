// Code generated by mockery v2.43.2. DO NOT EDIT.

package mocks

import (
	context "context"

	felt "github.com/NethermindEth/juno/core/felt"
	mock "github.com/stretchr/testify/mock"

	rpc "github.com/NethermindEth/starknet.go/rpc"

	starknet "github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
)

// Reader is an autogenerated mock type for the Reader type
type Reader struct {
	mock.Mock
}

// AccountNonce provides a mock function with given fields: _a0, _a1
func (_m *Reader) AccountNonce(_a0 context.Context, _a1 *felt.Felt) (*felt.Felt, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for AccountNonce")
	}

	var r0 *felt.Felt
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *felt.Felt) (*felt.Felt, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *felt.Felt) *felt.Felt); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*felt.Felt)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *felt.Felt) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// BlockWithTxHashes provides a mock function with given fields: ctx, blockID
func (_m *Reader) BlockWithTxHashes(ctx context.Context, blockID rpc.BlockID) (*rpc.Block, error) {
	ret := _m.Called(ctx, blockID)

	if len(ret) == 0 {
		panic("no return value specified for BlockWithTxHashes")
	}

	var r0 *rpc.Block
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, rpc.BlockID) (*rpc.Block, error)); ok {
		return rf(ctx, blockID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, rpc.BlockID) *rpc.Block); ok {
		r0 = rf(ctx, blockID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*rpc.Block)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, rpc.BlockID) error); ok {
		r1 = rf(ctx, blockID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Call provides a mock function with given fields: _a0, _a1, _a2
func (_m *Reader) Call(_a0 context.Context, _a1 rpc.FunctionCall, _a2 rpc.BlockID) ([]*felt.Felt, error) {
	ret := _m.Called(_a0, _a1, _a2)

	if len(ret) == 0 {
		panic("no return value specified for Call")
	}

	var r0 []*felt.Felt
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, rpc.FunctionCall, rpc.BlockID) ([]*felt.Felt, error)); ok {
		return rf(_a0, _a1, _a2)
	}
	if rf, ok := ret.Get(0).(func(context.Context, rpc.FunctionCall, rpc.BlockID) []*felt.Felt); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*felt.Felt)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, rpc.FunctionCall, rpc.BlockID) error); ok {
		r1 = rf(_a0, _a1, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CallContract provides a mock function with given fields: _a0, _a1
func (_m *Reader) CallContract(_a0 context.Context, _a1 starknet.CallOps) ([]*felt.Felt, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for CallContract")
	}

	var r0 []*felt.Felt
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, starknet.CallOps) ([]*felt.Felt, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, starknet.CallOps) []*felt.Felt); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*felt.Felt)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, starknet.CallOps) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Events provides a mock function with given fields: ctx, input
func (_m *Reader) Events(ctx context.Context, input rpc.EventsInput) (*rpc.EventChunk, error) {
	ret := _m.Called(ctx, input)

	if len(ret) == 0 {
		panic("no return value specified for Events")
	}

	var r0 *rpc.EventChunk
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, rpc.EventsInput) (*rpc.EventChunk, error)); ok {
		return rf(ctx, input)
	}
	if rf, ok := ret.Get(0).(func(context.Context, rpc.EventsInput) *rpc.EventChunk); ok {
		r0 = rf(ctx, input)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*rpc.EventChunk)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, rpc.EventsInput) error); ok {
		r1 = rf(ctx, input)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LatestBlockHeight provides a mock function with given fields: _a0
func (_m *Reader) LatestBlockHeight(_a0 context.Context) (uint64, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for LatestBlockHeight")
	}

	var r0 uint64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (uint64, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(context.Context) uint64); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TransactionByHash provides a mock function with given fields: _a0, _a1
func (_m *Reader) TransactionByHash(_a0 context.Context, _a1 *felt.Felt) (rpc.Transaction, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for TransactionByHash")
	}

	var r0 rpc.Transaction
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *felt.Felt) (rpc.Transaction, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *felt.Felt) rpc.Transaction); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(rpc.Transaction)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *felt.Felt) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TransactionReceipt provides a mock function with given fields: _a0, _a1
func (_m *Reader) TransactionReceipt(_a0 context.Context, _a1 *felt.Felt) (rpc.TransactionReceipt, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for TransactionReceipt")
	}

	var r0 rpc.TransactionReceipt
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *felt.Felt) (rpc.TransactionReceipt, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *felt.Felt) rpc.TransactionReceipt); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(rpc.TransactionReceipt)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *felt.Felt) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewReader creates a new instance of Reader. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *Reader {
	mock := &Reader{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
