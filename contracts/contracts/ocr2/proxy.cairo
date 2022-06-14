%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin, BitwiseBuiltin

from contracts.ocr2.interfaces.IAggregator import IAggregator

from contracts.ocr2.ownable import (
    Ownable_initializer,
    Ownable_only_owner,
    Ownable_get_owner,
    Ownable_transfer_ownership,
    Ownable_accept_ownership
)

struct Phase:
    member id: felt
    member aggregator: felt
end

@storage_var
func current_phase_() -> (phase: Phase):
end

@storage_var
func proposed_aggregator_() -> (address: felt):
end

@storage_var
func phases_(id: felt) -> (address: felt):
end

const SHIFT = 2 ** 128
const MAX_ID = SHIFT - 1

@constructor
func constructor{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}(
    owner: felt,
    address: felt
):
    Ownable_initializer(owner)
    set_aggregator(address)
    return ()
end

func set_aggregator{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}(
    address: felt
):
    let (current_phase: Phase) = current_phase_.read()
    let id = current_phase.id + 1
    current_phase_.write(Phase(id=id, aggregator=address))
    phases_.write(id, address)
    return ()
end

func propose_aggregator{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}(
    address: felt
):
    Ownable_only_owner()

    proposed_aggregator_.write(address)
    return ()
end

func confirm_aggregator{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}(
    address: felt
):
    Ownable_only_owner()

    let (proposed_aggregator) = proposed_aggregator_.read()
    assert proposed_aggregator = address
    proposed_aggregator_.write(0)
    set_aggregator(proposed_aggregator)
    return ()
end

@view
func latest_round_data{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}() -> (round: Round):
    # alloc_locals
    let (phase: Phase) = current_phase_.read()
    let (round: Round) = IAggregator.latest_round_data(contract_address=phase.aggregator)

    # Add phase_id to the high bits of round_id
    round.id = round.id + phase.id * SHIFT
    return (round)
end

@view
func round_data{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}(round_id: felt) -> (round: Round):
    let (phase, round_id) = split_felt(round_id)
    let (address) = phases_.read(phase)
    assert_not_zero(address)

    # Add phase_id to the high bits of round_id
    round.id = round.id + phase.id * SHIFT

    let (round: Round) = IAggregator.round_data(contract_address=address, round_id=round_id)
    return (round)
end

# These read from the proposed aggregator as a way to test the aggregator before making setting it live.

@view
func proposed_latest_round_data{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}() -> (round: Round):
    # alloc_locals
    let (aggregator) = proposed_aggregator_.read()
    let (round: Round) = IAggregator.latest_round_data(contract_address=aggregator)
    return (round)
end

@view
func proposed_round_data{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}(round_id: felt) -> (round: Round):
    let (aggregator) = proposed_aggregator_.read()
    let (round: Round) = IAggregator.round_data(contract_address=address, round_id=round_id)
    return (round)
end

@view
func aggregator{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}() -> (aggregator: felt):
    let (phase: Phase) = current_phase_.read()
    return (phase.aggregator)
end

@view
func phase_id{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}() -> (phase_id: felt):
    let (phase: Phase) = current_phase_.read()
    return (phase.id)
end

@view
func description{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}() -> (description: felt):
    let (phase: Phase) = current_phase_.read()
    let (description) = IAggregator.description(contract_address=phase.address)
    return (description)
end

@view
func decimals{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}() -> (decimals: felt):
    let (phase: Phase) = current_phase_.read()
    let (decimals) = IAggregator.decimals(contract_address=phase.address)
    return (decimals)
end

@view
func type_and_version{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}() -> (meta: felt):
    let (phase: Phase) = current_phase_.read()
    let (meta) = IAggregator.type_and_version(contract_address=phase.address)
    return (meta)
end
