%lang starknet

from starkware.cairo.common.alloc import alloc
from starkware.cairo.common.cairo_builtins import HashBuiltin
from starkware.cairo.common.math import assert_not_zero
from starkware.cairo.common.math_cmp import is_nn
from starkware.cairo.common.bool import TRUE, FALSE
from starkware.starknet.common.syscalls import get_block_timestamp

from chainlink.cairo.ocr2.IAggregator import IAggregator, Round

@storage_var
func PriceConsumerWithSequencerCheck_uptime_feed_address() -> (address : felt):
end

@storage_var
func PriceConsumerWithSequencerCheck_aggregator_address() -> (address : felt):
end

# If the sequencer is up and that 60 sec has passed,
# The function retrieves the latest price from the data feed using the priceFeed object.
@constructor
func constructor{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    uptime_feed_address : felt, aggregator_address : felt
):
    assert_not_zero(uptime_feed_address)
    assert_not_zero(aggregator_address)

    PriceConsumerWithSequencerCheck_uptime_feed_address.write(uptime_feed_address)
    PriceConsumerWithSequencerCheck_aggregator_address.write(aggregator_address)
    return ()
end

# If the sequencer is up and report is OK then we can get the latest price.
@view
func get_latest_price{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    round : felt
):
    assert_sequencer_healthy()
    let (aggregator_address) = PriceConsumerWithSequencerCheck_aggregator_address.read()
    let (round : Round) = IAggregator.latest_round_data(contract_address=aggregator_address)
    return (round=round.answer)
end

# Errors if the report is stale, or it's reported that Sequencer node is down
@external
func assert_sequencer_healthy{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}():
    alloc_locals

    let (uptime_feed_address) = PriceConsumerWithSequencerCheck_uptime_feed_address.read()

    # Get latest_round_data from sequencer contract which should be update by a message send from L1
    let (round : Round) = IAggregator.latest_round_data(contract_address=uptime_feed_address)
    let (local block_timestemp) = get_block_timestamp()
    let time = block_timestemp - round.updated_at

    # After 60 sec the report is considered stale
    let (is_ls) = is_nn(time - (60 + 1))

    # 0 if the sequencer is up and 1 if it is down
    if round.answer == 0:
        with_attr error_message("L2 Sequencer is up, report stale"):
            assert is_ls = 0
        end
        return ()
    end
    with_attr error_message("L2 Sequencer is down, report stale"):
        assert is_ls = 0
    end
    with_attr error_message("L2 Sequencer is down, report ok"):
        assert round.answer = 0
    end
    return ()
end
