%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin
from starkware.cairo.common.bool import TRUE, FALSE
from starkware.cairo.common.uint256 import Uint256
from starkware.starknet.common.syscalls import get_tx_info

@storage_var
func token677ReceiverMock_token_sender_() -> (address : felt):
end

@storage_var
func token677ReceiverMock_sent_value_() -> (value : Uint256):
end

@storage_var
func token677ReceiverMock_token_data_(index : felt) -> (data : felt):
end

@storage_var
func token677ReceiverMock_token_data_len_() -> (data_len : felt):
end

@storage_var
func token677ReceiverMock_called_fallback_() -> (bool : felt):
end

@constructor
func constructor{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}():
    token677ReceiverMock_called_fallback_.write(FALSE)
    return ()
end

@external
func onTokenTransfer{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    selector : felt, data_len : felt, data : felt*
):
    with_attr error_message("ERC677: address can not be null"):
        assert data_len = 3
    end
    let value : Uint256 = Uint256(low=data[0], high=data[1])
    token677ReceiverMock_called_fallback_.write(TRUE)
    let (tx_info) = get_tx_info()
    token677ReceiverMock_token_sender_.write(tx_info.account_contract_address)
    token677ReceiverMock_sent_value_.write(value)
    token677ReceiverMock_token_data_len_.write(data_len)
    fill_data_storage(0, data_len, data)
    return ()
end

func fill_data_storage{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    index : felt, data_len : felt, data : felt*
):
    if data_len == 0:
        return ()
    end

    let index = index + 1
    token677ReceiverMock_token_data_.write(index, [data])
    return fill_data_storage(index=index, data_len=data_len - 1, data=data + 1)
end

@view
func get_called_fallback{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    bool : felt
):
    let (bool) = token677ReceiverMock_called_fallback_.read()
    return (bool)
end

@view
func get_sent_value{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    value : Uint256
):
    let (value) = token677ReceiverMock_sent_value_.read()
    return (value)
end

@view
func get_token_sender{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    address : felt
):
    let (address) = token677ReceiverMock_token_sender_.read()
    return (address)
end
