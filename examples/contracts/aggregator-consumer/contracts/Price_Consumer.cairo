%lang starknet

from starkware.cairo.common.alloc import alloc
from starkware.cairo.common.cairo_builtins import HashBuiltin
from contracts.cairo.ocr2.interfaces.IAggregator import IAggregator, Round
from starkware.cairo.common.math import assert_not_zero
from starkware.cairo.common.math_cmp import is_nn
from starkware.cairo.common.bool import TRUE, FALSE
from starkware.starknet.common.syscalls import (
    get_block_timestamp,
)

@storage_var
func uptime_feed_address_() -> (address : felt):
end

@storage_var
func aggregator_address_() -> (address : felt):
end

# If the sequencer is up and that 60 sec has passed, 
# The function retrieves the latest price from the data feed using the priceFeed object.
@constructor
func constructor{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}(uptime_feed_address : felt, aggregator_address : felt):
    
    assert_not_zero(uptime_feed_address)
    assert_not_zero(aggregator_address)

    uptime_feed_address_.write(uptime_feed_address)
    aggregator_address_.write(aggregator_address)
    return ()
end

# If the sequencer is up and report is OK then we can get the latest price.
@view
func get_latest_price{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}() -> (round : felt):
    let (sequencer_is_up) = assert_sequencer_healthy()
    with_attr error_message("L2 Sequencer is down, report ok"):
        assert sequencer_is_up = TRUE
    end
    let (aggregator_address) = aggregator_address_.read()
    let (round : Round) = IAggregator.latest_round_data(contract_address=aggregator_address)
    return (round=round.answer)
end

# Check if the sequencer is up or down
@external
func assert_sequencer_healthy{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}() -> (sequencer_is_up : felt):
    alloc_locals

    let (uptime_feed_address) = uptime_feed_address_.read()

    # Get latest_round_data from sequencer contract which should be update by a message send from L1
    let (round : Round) = IAggregator.latest_round_data(contract_address=uptime_feed_address)
    let (local block_timestemp) = get_block_timestamp()

    # 0 if the sequencer is up and 1 if it is down
    let time = block_timestemp - round.transmission_timestamp
    
    # After 60 sec the report is considered stale
    let (is_ls) = is_nn(time - (60 + 1))
    if round.answer == 0:
        with_attr error_message("L2 Sequencer is up, report stale"):
            assert is_ls = 0
        end
        return (TRUE)
    end
    with_attr error_message("L2 Sequencer is down, report stale"):
        assert is_ls = 0
    end
    return (FALSE)
end
