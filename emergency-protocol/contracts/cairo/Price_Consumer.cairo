%lang starknet

from starkware.cairo.common.alloc import alloc
from starkware.cairo.common.cairo_builtins import HashBuiltin
from ocr2.interfaces.IAggregator import IAggregator, Round
from starkware.cairo.common.math import assert_not_zero
from starkware.cairo.common.math_cmp import is_nn
from starkware.cairo.common.bool import TRUE, FALSE
from starkware.starknet.common.syscalls import (
    get_block_timestamp,
)

from cairo.interfaces.IUptimeFeed import RoundFeed, IUptimeFeed

@storage_var
func uptime_feed_address_() -> (address : felt):
end

@storage_var
func aggregator_address_() -> (address : felt):
end

@storage_var
func get_round_() -> (round : RoundFeed):
end

@storage_var
func get_block_() -> (block : felt):
end

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


@view
func get_latest_price{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}() -> (round : felt):
    let (sequencer_is_up) = check_sequencer_state()
    with_attr error_message("L2 sequencer down: Chainlink feeds are not being updated"):
        assert sequencer_is_up = TRUE
    end
    let (aggregator_address) = aggregator_address_.read()
    let (round : Round) = IAggregator.latest_round_data(contract_address=aggregator_address)
    return (round.answer)
end

@external
func check_sequencer_state{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}() -> (sequencer_is_up : felt):
    alloc_locals

    let (uptime_feed_address) = uptime_feed_address_.read()
    let (round : RoundFeed) = IUptimeFeed.latest_round_data(contract_address=uptime_feed_address)
    let (local block_timestemp) = get_block_timestamp()
    get_round_.write(round)
    get_block_.write(block_timestemp - round.updated_at)
    if round.status == 0:
        let time = block_timestemp - round.updated_at
        let (is_ls) = is_nn(time - (3600 + 1))
        if is_ls == 0:
            return (TRUE)
        end
        return (FALSE)
    end

    return (FALSE)
end
