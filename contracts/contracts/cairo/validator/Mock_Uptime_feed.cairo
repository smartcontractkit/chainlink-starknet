%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin
from starkware.cairo.common.math_cmp import is_le
from starkware.starknet.common.syscalls import (
    get_block_timestamp,
)

from ocr2.interfaces.IAggregator import Round
from starkware.cairo.common.alloc import alloc

@storage_var
func l1_validator_() -> (address: felt):
end

@storage_var
func round_() -> (round: Round):
end

@constructor
func constructor{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}():
    return ()
end

# receive and handle messages from L1
@l1_handler
func update_status{pedersen_ptr : HashBuiltin*, syscall_ptr : felt*, range_check_ptr}(from_address : felt, status : felt, timestamp : felt):

    alloc_locals

    ### ADD ADDRESS ISTARKNETCORE + CHECK 
    let (l1_validator_address) = l1_validator_.read()
    assert from_address = l1_validator_address
    let (updated_at) = get_block_timestamp()
    round_.write(Round(round_id=1, answer=status, block_num=1, observation_timestamp=timestamp, transmission_timestamp=updated_at))
    return ()
end

@external
func set_l1_sender{pedersen_ptr : HashBuiltin*, syscall_ptr : felt*, range_check_ptr}(address : felt):
    l1_validator_.write(address)
    return ()
end

@view
func latest_round_data{pedersen_ptr : HashBuiltin*, syscall_ptr : felt*, range_check_ptr}() -> (round : Round):
    let (round) = round_.read()
    return (round)
end
