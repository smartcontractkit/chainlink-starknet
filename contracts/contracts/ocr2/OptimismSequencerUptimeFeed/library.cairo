%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin, SignatureBuiltin
from starkware.starknet.common.syscalls import get_tx_info, get_block_timestamp, get_caller_address
from starkware.starknet.common.math import assert_not_zero

from utils import assert_boolean

struct Round:
    member status : felt
    member started_at : felt
    member updated_at : felt
end

# TODO: maybe not needed
# struct FeedState:
#     member latest_round_id : felt
#     member latest_status : felt
#     member started_at : felt
#     member updated_at : felt
# end

# TODO: probably move to IAggregator
@event
func AnswerUpdated(current : felt, round_id : felt, timestamp : felt):
end

# TODO: probably move to IAggregator
@event
func NewRound(round_id : felt, started_by : felt, started_at : felt):
end

@event
func RoundUpdated(status : felt, updated_at : felt):
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
func s_rounds_len() -> (res : felt):
end

func require_l1_sender{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    sender : felt
):
    let (l1_sender) = s_l1_sender.read()
    let (sender) = get_caller_address()

    with_attr error_message("invalid sender"):
        assert l1_sender = sender
    end

    return ()
end

func require_valid_round_id(round_id : felt):
    let (lateset_round_id) = _get_latest_round_id()

    with_attr error_message("invalid round_id"):
        assert_not_zero(round_id)
        # TODO: do we need to check if uint80 is overflown?
        assert_le(round_id, lateset_round_id)
    end

    return ()
end

# TODO: maybe not needed
# @storage_var
# func s_feed_state() -> (res : FeedState):
# end

func _set_round{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    round_id : felt, value : Round
):
    s_rounds.write(id=round_id, field=Round.status, value=value.status)
    s_rounds.write(id=round_id, field=Round.started_at, value=value.started_at)
    s_rounds.write(id=round_id, field=Round.updated_at, value=value.updated_at)

    return ()
end

func _get_round{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    round_id : felt, value : Round
):
    let (status) = s_rounds.read(id=round_id, field=Round.status)
    let (started_at) = s_rounds.write(id=round_id, field=Round.started_at)
    let (updated_at) = s_rounds.write(id=round_id, field=Round.updated_at)

    return (Round(status, started_at, updated_at))
end

func _get_latest_status{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    res : felt
):
    let (latest_round_id) = s_rounds_len.read()
    let (latest_status) = s_rounds.read(latest_round_id, Round.status)

    return (latest_status)
end

func _get_latest_started_at{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    ) -> (res : felt):
    let (latest_round_id) = s_rounds_len.read()
    let (latest_started_at) = s_rounds.read(latest_round_id, Round.started_at)

    return (latest_started_at)
end

func _get_latest_round_id{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    res : felt
):
    let (latest_round_id) = s_rounds_len.read()

    return (latest_started_at)
end

func _set_l1_sender{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
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

func _record_round{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    round_id : felt, status : felt, timestamp : felt
):
    s_rounds_len.write(round_id)

    let (updated_at) = get_block_timestamp()
    let next_round = Round(status=status, started_at=timestamp, updated_at=updated_at)

    set_round(round_id, next_round)

    let (sender) = get_caller_address()
    NewRound.emit(round_id, sender, timestamp)
    AnswerUpdated.emit(status, round_id, timestamp)

    return ()
end

func _update_round{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    round_id : felt, status : felt
):
    let (updated_at) = get_block_timestamp()
    s_rounds.write(round_id, Round.uodated_at, updated_at)

    RoundUpdated.emit(status, updated_at)
    return ()
end

namespace optimism_sequencer_uptime_feed:
    func constructor{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
        l1_sender_address : felt,
        l2_cross_domain_messenger_addr : felt,
        initial_status : felt,
        owner_address : felt,
    ):
        assert_boolean(initial_status)

        simple_read_access_controller.constructor(owner_address)
        set_l1_sender(l1_sender_address)
        s_l2_cross_domain_messenger.write(l2_cross_domain_messenger_addr)

        # TODO: can not have uninitialized contracts
        # let feed_state = FeedState(latest_round_id=0, latest_status=FALSE, started_at=0, updated_at=0)
        # s_feed_state.write(feed_state)

        let (timestamp) = get_block_timestamp()
        record_round(1, initial_status, timestamp)

        return ()
    end

    func update_status{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
        status : felt, timestamp : felt
    ):
        require_l1_sender()
        assert_boolean(status)
    end

    func latest_answer{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
        res : felt
    ):
        simple_read_access_controller.check_access()

        let (answer) = _get_latest_status()
        return (answer)
    end

    func latest_timestamp{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
        res : felt
    ):
        simple_read_access_controller.check_access()

        let (timestamp) = _get_latest_started_at()
        return (timestamp)
    end

    func latest_round{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
        res : felt
    ):
        simple_read_access_controller.check_access()

        let (round) = _get_latest_round_id()
        return (round)
    end

    # Maybe unite with above?
    func get_answer{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
        round_id : felt
    ) -> (res : felt):
        simple_read_access_controller.check_access()
        require_valid_round_id(round_id)

        let (answer) = s_rounds.read(round_id, Round.status)
        return (answer)
    end

    func get_timestamp{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
        round_id : felt
    ) -> (res : felt):
        simple_read_access_controller.check_access()
        require_valid_round_id(round_id)

        let (timestamp) = s_rounds.read(round_id, Round.timestamp)
        return (timestamp)
    end

    # TODO: complete
    func get_round_data{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
        round_id : felt
    ) -> (res : felt):
        simple_read_access_controller.check_access()
        require_valid_round_id(round_id)

        let (round) = _get_round_data(round_id)
        return (round)
    end
end
