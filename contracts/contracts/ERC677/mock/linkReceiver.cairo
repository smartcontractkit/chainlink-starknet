%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin
from starkware.cairo.common.bool import TRUE, FALSE
from starkware.cairo.common.uint256 import Uint256
from starkware.starknet.common.syscalls import get_contract_address, library_call

from openzeppelin.token.erc20.IERC20 import IERC20

@storage_var
func linkReceiver_fallback_called_() -> (bool : felt):
end

@storage_var
func linkReceiver_call_data_called_() -> (bool : felt):
end

@storage_var
func linkReceiver_tokens_received_() -> (value : Uint256):
end

@storage_var
func linkReceiver_implementation_hash_() -> (class_hash : felt):
end

@constructor
func constructor{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    class_hash : felt
):
    linkReceiver_implementation_hash_.write(class_hash)
    return ()
end

@external
func onTokenTransfer{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    selector : felt, data_len : felt, data : felt*
):
    linkReceiver_fallback_called_.write(TRUE)
    let (class_hash) = linkReceiver_implementation_hash_.read()
    library_call(class_hash, selector, data_len, data)
    return ()
end

@external
func callbackWithoutWithdrawl{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}():
    linkReceiver_call_data_called_.write(TRUE)
    return ()
end

@external
func callbackWithWithdrawl{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    value_high : felt, value_low : felt, sender : felt, token_addr : felt
):
    let value : Uint256 = Uint256(low=value_low, high=value_high)
    linkReceiver_call_data_called_.write(TRUE)
    let (contract_address) = get_contract_address()
    IERC20.transferFrom(token_addr, sender, contract_address, value)

    linkReceiver_tokens_received_.write(value)
    return ()
end

@view
func get_fallback{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    bool : felt
):
    let (bool) = linkReceiver_fallback_called_.read()
    return (bool)
end

@view
func get_call_data{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    bool : felt
):
    let (bool) = linkReceiver_call_data_called_.read()
    return (bool)
end

@view
func get_tokens{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    value : Uint256
):
    let (value) = linkReceiver_tokens_received_.read()
    return (value)
end
