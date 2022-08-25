%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin
from starkware.cairo.common.bool import TRUE, FALSE
from starkware.cairo.common.math import split_felt 
from starkware.cairo.common.math_cmp import is_le
from starkware.cairo.common.uint256 import (
    Uint256,
    uint256_sub,
    uint256_mul,
    uint256_eq,
    uint256_signed_div_rem,
    uint256_cond_neg,
    uint256_le
)

const THRESHOLD_MULTIPLIER = 100000

from contracts.cairo.ownable import (
    Ownable_initializer,
    Ownable_only_owner,
)

@storage_var
func flags_() -> (address: felt):
end

@storage_var
func threshold_() -> (threshold: felt):
end

@constructor
func constructor{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}(
    owner: felt,
    flags: felt,
    threshold: felt,
):
    Ownable_initializer(owner)
    flags_.write(flags)
    threshold_.write(threshold)
    return ()
end

@external
func validate{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}(
    prev_round_id: felt,
    prev_answer: felt,
    round_id: felt,
    answer: felt
) -> (valid: felt):
    alloc_locals

    let (valid) = is_valid(prev_answer, answer)
    
    if valid == FALSE:
        # Do stuff
        let a = 1
    end

    return (valid)
end

# TODO: set_ events

@external
func set_flagging_threshold{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}(
    threshold: felt
):
    Ownable_only_owner()
    threshold_.write(threshold)
    return ()
end

@external
func set_flags_address{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}(
    flags: felt
):
    Ownable_only_owner()
    flags_.write(flags)
    return ()
end

# ---

func felt_to_uint256{range_check_ptr}(x) -> (uint_x : Uint256):
    let (high, low) = split_felt(x)
    return (Uint256(low=low, high=high))
end

# TODO: quadruple test the logic in this method to ensure it can never fail & revert
func is_valid{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}(
    _prev_answer: felt,
    _answer: felt
) -> (valid: felt):
    alloc_locals

    if _prev_answer == 0:
        # TODO: I'd rather check round_id
        return (valid=TRUE)
    end
    
    let (prev_answer: Uint256) = felt_to_uint256(_prev_answer)
    let (answer: Uint256) = felt_to_uint256(_answer)
    
    # TODO: how is underflow/overflow handled here?
    let (change: Uint256) = uint256_sub(prev_answer, answer)
    let (multiplier: Uint256) = felt_to_uint256(THRESHOLD_MULTIPLIER)

    let (numerator: Uint256, overflow: Uint256) = uint256_mul(change, multiplier)
    let (zero) = uint256_eq(overflow, Uint256(0, 0))    
    # If overflow is not zero then we overflowed
    if zero != TRUE:
        return (valid=FALSE)
    end

    let (ratio: Uint256, _remainder) = uint256_signed_div_rem(numerator, prev_answer)

    # Take the absolute value of ratio.
    # https://github.com/starkware-libs/cairo-lang/blob/b614d1867c64f3fb2cf4a4879348cfcf87c3a5a7/src/starkware/cairo/common/uint256.cairo#L261-L264=
    let (local ratio_sign) = is_le(2 ** 127, ratio.high)
    local range_check_ptr = range_check_ptr
    let (local ratio) = uint256_cond_neg(ratio, should_neg=ratio_sign)
    # TODO: can it be simplified via sign()?

    let (threshold_felt) = threshold_.read()
    let (threshold: Uint256) = felt_to_uint256(threshold_felt)
    # ratio <= threshold
    let (is_le_) = uint256_le(ratio, threshold)
    return (valid=is_le_)
end