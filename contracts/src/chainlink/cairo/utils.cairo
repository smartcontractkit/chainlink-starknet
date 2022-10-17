from starkware.cairo.common.math import assert_in_range, assert_lt_felt
from starkware.cairo.common.uint256 import Uint256
from starkware.cairo.common.math import split_felt

func assert_boolean{range_check_ptr}(value : felt):
    # Asserts that value is in the range [0, 2).
    return assert_in_range(value, 0, 2)
end

func felt_to_uint256{range_check_ptr}(x) -> (uint_x : Uint256):
    let (high, low) = split_felt(x)
    return (Uint256(low=low, high=high))
end

func uint256_to_felt{range_check_ptr}(value : Uint256) -> (value : felt):
    assert_lt_felt(value.high, 2 ** 123)
    return (value.high * (2 ** 128) + value.low)
end
