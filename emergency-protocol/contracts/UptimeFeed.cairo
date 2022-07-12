%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin
from starkware.cairo.common.math_cmp import is_le
from starkware.starknet.common.syscalls import (
    get_block_timestamp,
)
from starkware.cairo.common.alloc import alloc

struct FeedState:
    member latest_round_id : felt
    member latest_status : felt
    member started_at : felt
    member updated_at : felt
end

@storage_var
func l1_validator_() -> (address: felt):
end

@storage_var
func feed_state_() -> (feedState: FeedState):
end

@event
func UpdateIgnored(
    latest_status: felt,
    latest_timestamp: felt,
    incoming_status: felt,
    incoming_timestamp: felt,
):
end


@constructor
func constructor{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}(l1_validator_address : felt):
    l1_validator_.write(l1_validator_address)
    return ()
end

# receive and handle messages from L1
@l1_handler
func updateStatus{pedersen_ptr : HashBuiltin*, syscall_ptr : felt*, range_check_ptr}(from_address : felt, status : felt, timestamp : felt):

    alloc_locals

    let (l1_validator_address) = l1_validator_.read()
    assert from_address = l1_validator_address
    
    # Ignore if latest recorded timestamp is newer
    let (feed_state) = feed_state_.read()

    let (is_le_) = is_le(timestamp, feed_state.started_at)

    if is_le_ == 1:
        UpdateIgnored.emit(feed_state.latest_status, feed_state.started_at, status, timestamp)
        return ()
    end

    let (updated_at) = get_block_timestamp()
    if feed_state.latest_status == status:
        feed_state_.write(FeedState(
            latest_round_id=feed_state.latest_round_id,
            latest_status=feed_state.latest_status,
            started_at=feed_state.started_at,
            updated_at=updated_at
        ))
    else:
        let latest_round_id = feed_state.latest_round_id + 1
        feed_state_.write(FeedState(
            latest_round_id=latest_round_id,
            latest_status=status,
            started_at=timestamp,
            updated_at=updated_at
        ))
    end
    return ()
end
