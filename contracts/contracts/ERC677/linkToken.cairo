%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin
from starkware.cairo.common.alloc import alloc
from starkware.cairo.common.math import assert_not_zero, split_felt
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

from contracts.ERC677.ERC677 import (
    initialize,
    transfer,
    transferFrom,
    approve,
    increaseAllowance,
    decreaseAllowance,
    permissionedMint,
    permissionedBurn,
    transferAndCall,
)

const NAME = 'ChainLink Token'
const SYMBOL = 'LINK'

@event
func Transfer(from_ : felt, to : felt, value : Uint256, data_len : felt, data : felt*):
end

@constructor
func constructor{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(minter : felt):
    alloc_locals

    let (init_vector : felt*) = alloc()
    assert init_vector[0] = NAME
    assert init_vector[1] = SYMBOL
    assert init_vector[2] = 18
    assert init_vector[3] = minter
    initialize(4, init_vector)
    return ()
end

@view
func type_and_version() -> (meta : felt):
    return ('LinkToken 0.0.1')
end
