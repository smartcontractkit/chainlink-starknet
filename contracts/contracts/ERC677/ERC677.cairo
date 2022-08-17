%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin
from starkware.cairo.common.alloc import alloc
from starkware.cairo.common.math import assert_not_zero, assert_le
from starkware.cairo.common.uint256 import (
    Uint256,
    uint256_add,
    uint256_check,
    uint256_le,
    uint256_lt,
    uint256_sub,
)
from starkware.starknet.common.syscalls import get_caller_address
from starkware.cairo.common.bool import FALSE, TRUE
from contracts.ERC677.ERC20.ERC20_base import (
    ERC20_allowances,
    ERC20_approve,
    ERC20_burn,
    ERC20_initializer,
    ERC20_mint,
    ERC20_transfer,
    allowance,
    balanceOf,
    decimals,
    name,
    symbol,
    totalSupply,
)
from contracts.ERC677.ERC20.permitted import (
    permitted_initializer,
    permitted_minter,
    permitted_minter_only,
    permittedMinter,
)
from contracts.ERC677.ERC20.initializable import initialized, set_initialized
from contracts.ERC677.interfaces.IERC677Receiver import IERC677Receiver

@event
func Transfer(from_ : felt, to : felt, value : Uint256, data_len : felt, data : felt*):
end

# Constructor (as initializer).

@external
func initialize{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    init_vector_len : felt, init_vector : felt*
):
    set_initialized()
    # We expect the init vector to be [name , symbol , decimals , minter_address].
    with_attr error_message("ILLEGAL_INIT_SIZE"):
        assert init_vector_len = 4
    end

    let name = [init_vector]
    let symbol = [init_vector + 1]
    let decimals = [init_vector + 2]
    ERC20_initializer(name, symbol, decimals)

    let minter_address = [init_vector + 3]
    permitted_initializer(minter_address)
    return ()
end

# Externals.

@external
func transfer{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    recipient : felt, amount : Uint256
) -> (success : felt):
    let (sender) = get_caller_address()
    ERC20_transfer(sender, recipient, amount)

    # Cairo equivalent to 'return (true)'
    return (1)
end

@external
func transferFrom{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    sender : felt, recipient : felt, amount : Uint256
) -> (success : felt):
    alloc_locals
    let (local caller) = get_caller_address()
    let (local caller_allowance : Uint256) = ERC20_allowances.read(owner=sender, spender=caller)

    # Validates amount <= caller_allowance and returns 1 if true.
    let (enough_allowance) = uint256_le(amount, caller_allowance)
    assert_not_zero(enough_allowance)

    ERC20_transfer(sender, recipient, amount)

    # Subtract allowance.
    let (new_allowance : Uint256) = uint256_sub(caller_allowance, amount)
    ERC20_allowances.write(sender, caller, new_allowance)

    # Cairo equivalent to 'return (true)'
    return (1)
end

@external
func approve{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    spender : felt, amount : Uint256
) -> (success : felt):
    let (caller) = get_caller_address()
    ERC20_approve(caller, spender, amount)

    # Cairo equivalent to 'return (true)'
    return (1)
end

@external
func increaseAllowance{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    spender : felt, added_value : Uint256
) -> (success : felt):
    alloc_locals
    uint256_check(added_value)
    let (local caller) = get_caller_address()
    let (local current_allowance : Uint256) = ERC20_allowances.read(caller, spender)

    # Add allowance.
    let (local new_allowance : Uint256, is_overflow) = uint256_add(current_allowance, added_value)
    assert (is_overflow) = 0

    ERC20_approve(caller, spender, new_allowance)

    # Cairo equivalent to 'return (true)'
    return (1)
end

@external
func decreaseAllowance{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    spender : felt, subtracted_value : Uint256
) -> (success : felt):
    alloc_locals
    uint256_check(subtracted_value)
    let (local caller) = get_caller_address()
    let (local current_allowance : Uint256) = ERC20_allowances.read(owner=caller, spender=spender)
    let (local new_allowance : Uint256) = uint256_sub(current_allowance, subtracted_value)

    # Validates new_allowance < current_allowance and returns 1 if true.
    let (enough_allowance) = uint256_lt(new_allowance, current_allowance)
    assert_not_zero(enough_allowance)

    ERC20_approve(caller, spender, new_allowance)

    # Cairo equivalent to 'return (true)'
    return (1)
end

@external
func permissionedMint{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    recipient : felt, amount : Uint256
):
    alloc_locals
    permitted_minter_only()
    local syscall_ptr : felt* = syscall_ptr

    ERC20_mint(recipient=recipient, amount=amount)

    return ()
end

@external
func permissionedBurn{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    account : felt, amount : Uint256
):
    alloc_locals
    permitted_minter_only()
    local syscall_ptr : felt* = syscall_ptr

    ERC20_burn(account=account, amount=amount)

    return ()
end

@external
func transferAndCall{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    to : felt, value : Uint256, selector : felt, data_len : felt, data : felt*
) -> (success : felt):
    alloc_locals

    let (caller) = get_caller_address()
    with_attr error_message("ERC677: address can not be null"):
        assert_not_zero(to)
    end
    ERC20_transfer(caller, to, value)
    Transfer.emit(caller, to, value, data_len, data)

    contractFallback(to, selector, data_len, data)
    return (TRUE)
end

# PRIVATE

func contractFallback{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    to : felt, selector : felt, data_len : felt, data : felt*
):
    IERC677Receiver.onTokenTransfer(to, selector, data_len, data)
    return ()
end

# func fill_data{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
#     index : felt, data_len : felt, data : felt*
# ) -> (len : felt):
#     if data_len == 0:
#         return ()
#     end

# let index = index + 1
#     token_data_.write(index, [data])
#     return fill_data_storage(index=index, data_len=data_len - 1, data=data + 1)
# end
