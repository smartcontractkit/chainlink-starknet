%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin, SignatureBuiltin
from starkware.starknet.common.syscalls import get_tx_info, get_block_timestamp
from starkware.cairo.common.bool import TRUE, FALSE

from ocr2.OptimismSequencerUptimeFeed.library import (
    set_l1_sender,
    s_l2_cross_domain_messenger,
    s_feed_state,
    record_round,
    FeedState,
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
    simple_read_access_controller.constructor(owner_address)
    set_l1_sender(l1_sender_address)
    s_l2_cross_domain_messenger.write(l2_cross_domain_messenger_addr)
    let feed_state = FeedState(latest_round_id=0, latest_status=FALSE, started_at=0, updated_at=0)
    s_feed_state.write(feed_state)

    let (timestamp) = get_block_timestamp()
    record_round(1, initial_status, timestamp)

    return ()
end
