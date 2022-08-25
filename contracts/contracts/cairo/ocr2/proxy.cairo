%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin, BitwiseBuiltin
from starkware.cairo.common.math import split_felt, assert_not_zero

from contracts.cairo.ocr2.interfaces.IAggregator import IAggregator, Round

from contracts.cairo.ownable import (
    Ownable_initializer,
    Ownable_only_owner,
    Ownable_get_owner,
    Ownable_transfer_ownership,
    Ownable_accept_ownership,
)

struct Phase:
    member id : felt
    member aggregator : felt
end

@storage_var
func current_phase_() -> (phase : Phase):
end

@storage_var
func proposed_aggregator_() -> (address : felt):
end

@storage_var
func phases_(id : felt) -> (address : felt):
end

const SHIFT = 2 ** 128
const MAX_ID = SHIFT - 1

@event
func aggregator_proposed(current : felt, proposed : felt):
end

@event
func aggregator_confirmed(previous : felt, latest : felt):
end

@constructor
func constructor{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    owner : felt, address : felt
):
    Ownable_initializer(owner)
    set_aggregator(address)
    return ()
end

func set_aggregator{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    address : felt
):
    let (current_phase : Phase) = current_phase_.read()
    let id = current_phase.id + 1
    current_phase_.write(Phase(id=id, aggregator=address))
    phases_.write(id, address)
    return ()
end

@external
func propose_aggregator{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    address : felt
):
    Ownable_only_owner()

    proposed_aggregator_.write(address)

    # emit event
    let (phase : Phase) = current_phase_.read()
    aggregator_proposed.emit(current=phase.aggregator, proposed=address)
    return ()
end

@external
func confirm_aggregator{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    address : felt
):
    Ownable_only_owner()

    let (phase : Phase) = current_phase_.read()
    let previous = phase.aggregator

    let (proposed_aggregator) = proposed_aggregator_.read()
    assert proposed_aggregator = address
    proposed_aggregator_.write(0)
    set_aggregator(proposed_aggregator)

    # emit event
    aggregator_confirmed.emit(previous=previous, latest=address)
    return ()
end

@view
func latest_round_data{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    round : Round
):
    # alloc_locals
    let (phase : Phase) = current_phase_.read()
    let (round : Round) = IAggregator.latest_round_data(contract_address=phase.aggregator)

    # Add phase_id to the high bits of round_id
    let round_id = round.round_id + (phase.id * SHIFT)
    return (
        Round(
        round_id=round_id,
        answer=round.answer,
        block_num=round.block_num,
        started_at=round.started_at,
        updated_at=round.updated_at,
        ),
    )
end

@view
func round_data{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    round_id : felt
) -> (round : Round):
    let (phase_id, round_id) = split_felt(round_id)
    let (address) = phases_.read(phase_id)
    assert_not_zero(address)

    let (round : Round) = IAggregator.round_data(contract_address=address, round_id=round_id)
    # Add phase_id to the high bits of round_id
    let round_id = round.round_id + (phase_id * SHIFT)
    return (
        Round(
        round_id=round_id,
        answer=round.answer,
        block_num=round.block_num,
        started_at=round.started_at,
        updated_at=round.updated_at,
        ),
    )
end

# These read from the proposed aggregator as a way to test the aggregator before making setting it live.

@view
func proposed_latest_round_data{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    ) -> (round : Round):
    # alloc_locals
    let (aggregator) = proposed_aggregator_.read()
    let (round : Round) = IAggregator.latest_round_data(contract_address=aggregator)
    return (round)
end

@view
func proposed_round_data{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    round_id : felt
) -> (round : Round):
    let (aggregator) = proposed_aggregator_.read()
    let (round : Round) = IAggregator.round_data(contract_address=aggregator, round_id=round_id)
    return (round)
end

@view
func aggregator{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    aggregator : felt
):
    let (phase : Phase) = current_phase_.read()
    return (phase.aggregator)
end

@view
func phase_id{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    phase_id : felt
):
    let (phase : Phase) = current_phase_.read()
    return (phase.id)
end

@view
func description{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    description : felt
):
    let (phase : Phase) = current_phase_.read()
    let (description) = IAggregator.description(contract_address=phase.aggregator)
    return (description)
end

@view
func decimals{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    decimals : felt
):
    let (phase : Phase) = current_phase_.read()
    let (decimals) = IAggregator.decimals(contract_address=phase.aggregator)
    return (decimals)
end

@view
func type_and_version{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    meta : felt
):
    let (phase : Phase) = current_phase_.read()
    let (meta) = IAggregator.type_and_version(contract_address=phase.aggregator)
    return (meta)
end
