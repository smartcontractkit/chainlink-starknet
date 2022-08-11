%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin, SignatureBuiltin
from starkware.starknet.common.syscalls import (
    get_tx_info,
    get_block_timestamp,
    get_caller_address,
    get_block_number,
)
from starkware.cairo.common.math import assert_not_zero, assert_le
from starkware.cairo.common.math_cmp import is_le
from starkware.cairo.common.bool import TRUE, FALSE

from cairo.utils import assert_boolean
from cairo.ocr2.interfaces.IAggregator import Round, AnswerUpdated, NewRound
from cairo.SimpleReadAccessController.library import simple_read_access_controller
from cairo.ownable import Ownable_only_owner

@event
func RoundUpdated(status : felt, transmission_timestamp : felt):
end

@event
func UpdateIgnored(
    latest_status : felt, latest_timestamp : felt, incoming_status : felt, incoming_timestamp : felt
):
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
func s_latest_round_id() -> (res : felt):
end

func require_l1_sender{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    address : felt
):
    let (l1_sender) = s_l1_sender.read()
    with_attr error_message("invalid sender"):
        assert l1_sender = address
    end

    return ()
end

func require_valid_round_id{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    round_id : felt
):
    let (latest_round_id) = s_latest_round_id.read()

    with_attr error_message("invalid round_id"):
        assert_not_zero(round_id)
        # TODO: do we need to check if uint80 is overflown?
        assert_le(round_id, latest_round_id)
    end

    return ()
end

func require_access{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}():
    let (address) = get_caller_address()
    simple_read_access_controller.check_access(address)

    return ()
end

# TODO: this methods are virtual in .sol: clarify
@external
func set_l1_sender{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    address : felt
):
    Ownable_only_owner()
    _set_l1_sender(address)

    return ()
end

@view
func l1_sender{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    address : felt
):
    let (address) = s_l1_sender.read()
    return (address)
end

namespace sequencer_uptime_feed:
    func constructor{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
        initial_status : felt, owner_address : felt
    ):
        assert_boolean(initial_status)

        simple_read_access_controller.constructor(owner_address)

        let (timestamp) = get_block_timestamp()
        _record_round(1, initial_status, timestamp)

        return ()
    end

    func update_status{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
        from_address : felt, status : felt, timestamp : felt
    ):
        alloc_locals
        require_l1_sender(from_address)
        assert_boolean(status)

        let (latest_round_id) = s_latest_round_id.read()
        let (latest_observation_timestamp) = s_rounds.read(
            latest_round_id, Round.observation_timestamp
        )
        let (local latest_status) = s_rounds.read(latest_round_id, Round.answer)

        let (lt) = is_le(timestamp, latest_observation_timestamp - 1)  # timestamp < latest_observation_timestamp
        if lt == TRUE:
            UpdateIgnored.emit(latest_status, latest_observation_timestamp, status, timestamp)
            return ()
        end

        if latest_status == status:
            _update_round(latest_round_id, status)
        else:
            let latest_round_id = latest_round_id + 1
            _record_round(latest_round_id, status, timestamp)
        end

        return ()
    end

    func latest_round_data{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
        round : Round
    ):
        require_access()

        let (latest_round_id) = s_latest_round_id.read()
        let (latest_round) = sequencer_uptime_feed.round_data(latest_round_id)

        return (latest_round)
    end

    func round_data{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
        round_id : felt
    ) -> (res : Round):
        require_access()
        require_valid_round_id(round_id)

        let (round) = _get_round(round_id)
        return (round)
    end

    func description{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
        description : felt
    ):
        const description = 'L2 Sequencer Uptime Status Feed'
        return (description)
    end

    func decimals{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
        decimals : felt
    ):
        const decimals = 0
        return (decimals)
    end

    func type_and_version{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
        meta : felt
    ):
        const meta = 'SequencerUptimeFeed 1.0.0'
        return (meta)
    end
end

func _set_round{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    round_id : felt, value : Round
):
    s_rounds.write(id=round_id, field=Round.answer, value=value.answer)
    s_rounds.write(id=round_id, field=Round.block_num, value=value.block_num)
    s_rounds.write(
        id=round_id, field=Round.observation_timestamp, value=value.observation_timestamp
    )
    s_rounds.write(
        id=round_id, field=Round.transmission_timestamp, value=value.transmission_timestamp
    )

    return ()
end

func _get_round{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    round_id : felt
) -> (round : Round):
    let (answer) = s_rounds.read(id=round_id, field=Round.answer)
    let (block_num) = s_rounds.read(id=round_id, field=Round.block_num)
    let (observation_timestamp) = s_rounds.read(id=round_id, field=Round.observation_timestamp)
    let (transmission_timestamp) = s_rounds.read(id=round_id, field=Round.transmission_timestamp)

    return (Round(round_id, answer, block_num, observation_timestamp, transmission_timestamp))
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
    s_latest_round_id.write(round_id)

    let (block_num) = get_block_number()
    let (transmission_timestamp) = get_block_timestamp()

    let round = Round(
        round_id=round_id,
        answer=status,
        block_num=block_num,
        observation_timestamp=timestamp,
        transmission_timestamp=transmission_timestamp,
    )
    _set_round(round_id, round)

    let (sender) = get_caller_address()
    NewRound.emit(round)
    AnswerUpdated.emit(status, round_id, timestamp)

    return ()
end

func _update_round{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    round_id : felt, status : felt
):
    let (transmission_timestamp) = get_block_timestamp()
    s_rounds.write(round_id, Round.transmission_timestamp, transmission_timestamp)

    RoundUpdated.emit(status, transmission_timestamp)
    return ()
end
