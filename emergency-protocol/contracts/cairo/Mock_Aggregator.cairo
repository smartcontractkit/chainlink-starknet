%lang starknet

from starkware.cairo.common.alloc import alloc
from starkware.cairo.common.cairo_builtins import HashBuiltin
from ocr2.interfaces.IAggregator import Round

@constructor
func constructor{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}():
    return ()
end

@view
func latest_round_data{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}() -> (round: Round):
    alloc_locals
   
    let round = Round(
        round_id=1,
        answer=3,
        block_num=2,
        observation_timestamp=56,
        transmission_timestamp=42,
    )
    return (round)
end
