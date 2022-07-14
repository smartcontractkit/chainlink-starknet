from starkware.cairo.common.math import assert_in_range

func assert_boolean{range_check_ptr}(value : felt):
    const lower_bound = 0
    const upper_bound = 2
    with_attr error_message("value isn't a boolean"):
        assert_in_range(value, lower_bound, upper_bound)
    end

    return ()
end
