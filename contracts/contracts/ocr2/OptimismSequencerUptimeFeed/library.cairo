%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin, SignatureBuiltin
from starkware.starknet.common.syscalls import get_tx_info, get_block_timestamp, get_caller_address

struct Round:
    member status : felt
    member started_at : felt
    member updated_at : felt
end

struct FeedState:
    member latest_round_id : felt
    member latest_status : felt
    member started_at : felt
    member updated_at : felt
end

# TODO: probably move to IAggregator
@event
func AnswerUpdated(current : felt, round_id : felt, timestamp : felt):
end

# TODO: probably move to IAggregator
@event
func NewRound(round_id : felt, started_by : felt, started_at : felt):
end

@event
func L1SenderTransferred(prev : felt, cur : felt):
end

@storage_var
func s_l1_sender() -> (address : felt):
end

@storage_var
func s_l2_cross_domain_messenger() -> (address : felt):
end

@storage_var
func s_rounds(id : felt, field : felt) -> (res : felt):
end

@storage_var
func s_feed_state() -> (res : FeedState):
end

func set_round{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    round_id : felt, value : Round
):
    s_rounds.write(id=round_id, field=Round.status, value=value.status)
    s_rounds.write(id=round_id, field=Round.started_at, value=value.started_at)
    s_rounds.write(id=round_id, field=Round.updated_at, value=value.updated_at)

    return ()
end

func set_l1_sender{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    address : felt
):
    let (old_address) = s_l1_sender.read()

    if old_address != address:
        s_l1_sender.write(address)
        L1SenderTransferred.emit(old_address, address)
        return ()
    end

    return ()
end

func record_round{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    round_id : felt, status : felt, timestamp : felt
):
    let (updated_at) = get_block_timestamp()
    let next_round = Round(status=status, started_at=timestamp, updated_at=updated_at)
    let feed_state = FeedState(
        latest_round_id=round_id, latest_status=status, started_at=timestamp, updated_at=updated_at
    )

    set_round(round_id, next_round)
    s_feed_state.write(feed_state)

    let (sender) = get_caller_address()
    NewRound.emit(round_id, sender, timestamp)
    AnswerUpdated.emit(status, round_id, timestamp)

    return ()
end
