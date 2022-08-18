%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin
from starkware.cairo.common.alloc import alloc
from contracts.ERC677.interfaces.IERC677Receiver import IERC677Receiver

from contracts.ERC677.starkgate_token import (
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

const NAME = 'Example ERC677 Token'
const SYMBOL = 'ERC677'

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
