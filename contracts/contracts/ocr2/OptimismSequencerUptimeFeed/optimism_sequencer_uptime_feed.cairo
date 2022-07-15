%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin, SignatureBuiltin
from starkware.starknet.common.syscalls import get_tx_info, get_block_timestamp
from starkware.cairo.common.bool import TRUE, FALSE

from ocr2.OptimismSequencerUptimeFeed.library import (
    set_l1_sender,
    s_l2_cross_domain_messenger,
    s_feed_state,
    record_round,
    optimism_sequencer_uptime_feed,
)
from ocr2.interfaces.IAggregator import IAggregator
from ocr2.interfaces.IAccessController import IAccessController
from ocr2.interfaces.IOptimismSequencerUptimeFeed import IOptimismSequencerUptimeFeed
from SimpleReadAccessController.library import simple_read_access_controller

@constructor
func constructor{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    l1_sender_address : felt,
    l2_cross_domain_messenger_addr : felt,
    initial_status : felt,
    owner_address : felt,
):
    optimism_sequencer_uptime_feed.constructor(
        l1_sender_address, l2_cross_domain_messenger_addr, initial_status, owner_address
    )
    return ()
end

@external
func update_status{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    status : felt, timestamp : felt
):
    optimism_sequencer_uptime_feed.update_status(status, timestamp)
    return ()
end

# TODO: clarify if shall be added to some interface
@view
func latest_answer{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    res : felt
):
    let (answer) = simple_read_access_controller.latest_answer()
    return (answer)
end

@view
func latest_timestamp{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    res : felt
):
    let (timestamp) = simple_read_access_controller.latest_timestamp()
    return (timestamp)
end

@view
func latest_round{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    res : felt
):
    let (round) = simple_read_access_controller.latest_round()
    return (round)
end

@view
func get_answer{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    round_id : felt
) -> (res : felt):
    let (answer) = simple_read_access_controller.get_answer(round_id)
    return (answer)
end

@view
func get_timestamp{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    round_id : felt
) -> (res : felt):
    let (timestamp) = simple_read_access_controller.get_timestamp(round_id)
    return (timestamp)
end
