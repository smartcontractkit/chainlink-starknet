// amarna: disable=arithmetic-div,arithmetic-sub,arithmetic-mul,arithmetic-add
%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin
from starkware.cairo.common.math import split_felt, assert_not_zero
from starkware.starknet.common.syscalls import get_caller_address

from chainlink.cairo.ocr2.IAggregator import IAggregator, Round

from chainlink.cairo.access.SimpleReadAccessController.library import SimpleReadAccessController
from chainlink.cairo.access.SimpleWriteAccessController.library import (
    owner,
    proposed_owner,
    transfer_ownership,
    accept_ownership,
    add_access,
    remove_access,
    enable_access_check,
    disable_access_check,
)
from chainlink.cairo.access.ownable import Ownable

struct Phase {
    id: felt,
    aggregator: felt,
}

@storage_var
func AggregatorProxy_current_phase() -> (phase: Phase) {
}

@storage_var
func AggregatorProxy_proposed_aggregator() -> (address: felt) {
}

@storage_var
func AggregatorProxy_phases(id: felt) -> (address: felt) {
}

const SHIFT = 2 ** 128;
const MAX_ID = SHIFT - 1;

@event
func AggregatorProposed(current: felt, proposed: felt) {
}

@event
func AggregatorConfirmed(previous: felt, latest: felt) {
}

@constructor
func constructor{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}(
    owner: felt, address: felt
) {
    SimpleReadAccessController.initialize(owner);  // This also calls Ownable.initializer
    set_aggregator(address);
    return ();
}

func set_aggregator{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}(
    address: felt
) {
    let (current_phase: Phase) = AggregatorProxy_current_phase.read();
    let id = current_phase.id + 1;
    AggregatorProxy_current_phase.write(Phase(id=id, aggregator=address));
    AggregatorProxy_phases.write(id, address);
    return ();
}

@external
func propose_aggregator{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}(
    address: felt
) {
    Ownable.assert_only_owner();
    with_attr error_message("AggregatorProxy: aggregator is zero address") {
        assert_not_zero(address);
    }
    AggregatorProxy_proposed_aggregator.write(address);

    // emit event
    let (phase: Phase) = AggregatorProxy_current_phase.read();
    AggregatorProposed.emit(current=phase.aggregator, proposed=address);
    return ();
}

@external
func confirm_aggregator{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}(
    address: felt
) {
    Ownable.assert_only_owner();
    with_attr error_message("AggregatorProxy: aggregator is zero address") {
        assert_not_zero(address);
    }
    let (phase: Phase) = AggregatorProxy_current_phase.read();
    let previous = phase.aggregator;

    let (proposed_aggregator) = AggregatorProxy_proposed_aggregator.read();
    assert proposed_aggregator = address;
    AggregatorProxy_proposed_aggregator.write(0);
    set_aggregator(proposed_aggregator);

    // emit event
    AggregatorConfirmed.emit(previous=previous, latest=address);
    return ();
}

// Read access helper
func require_access{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}() {
    let (address) = get_caller_address();
    SimpleReadAccessController.check_access(address);

    return ();
}

@view
func latest_round_data{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}() -> (
    round: Round
) {
    require_access();
    let (phase: Phase) = AggregatorProxy_current_phase.read();
    let (round: Round) = IAggregator.latest_round_data(contract_address=phase.aggregator);

    // Add phase_id to the high bits of round_id
    let round_id = round.round_id + (phase.id * SHIFT);
    return (
        Round(
        round_id=round_id,
        answer=round.answer,
        block_num=round.block_num,
        started_at=round.started_at,
        updated_at=round.updated_at,
        ),
    );
}

@view
func round_data{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}(
    round_id: felt
) -> (round: Round) {
    require_access();
    let (phase_id, round_id) = split_felt(round_id);
    let (address) = AggregatorProxy_phases.read(phase_id);
    assert_not_zero(address);

    let (round: Round) = IAggregator.round_data(contract_address=address, round_id=round_id);
    // Add phase_id to the high bits of round_id
    let round_id = round.round_id + (phase_id * SHIFT);
    return (
        Round(
        round_id=round_id,
        answer=round.answer,
        block_num=round.block_num,
        started_at=round.started_at,
        updated_at=round.updated_at,
        ),
    );
}

// These read from the proposed aggregator as a way to test the aggregator before making setting it live.

@view
func proposed_latest_round_data{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}(
    ) -> (round: Round) {
    require_access();
    let (aggregator) = AggregatorProxy_proposed_aggregator.read();
    let (round: Round) = IAggregator.latest_round_data(contract_address=aggregator);
    return (round,);
}

@view
func proposed_round_data{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}(
    round_id: felt
) -> (round: Round) {
    require_access();
    let (aggregator) = AggregatorProxy_proposed_aggregator.read();
    let (round: Round) = IAggregator.round_data(contract_address=aggregator, round_id=round_id);
    return (round,);
}

@view
func aggregator{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}() -> (
    aggregator: felt
) {
    require_access();
    let (phase: Phase) = AggregatorProxy_current_phase.read();
    return (phase.aggregator,);
}

@view
func phase_id{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}() -> (
    phase_id: felt
) {
    require_access();
    let (phase: Phase) = AggregatorProxy_current_phase.read();
    return (phase.id,);
}

@view
func description{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}() -> (
    description: felt
) {
    require_access();
    let (phase: Phase) = AggregatorProxy_current_phase.read();
    let (description) = IAggregator.description(contract_address=phase.aggregator);
    return (description,);
}

@view
func decimals{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}() -> (
    decimals: felt
) {
    require_access();
    let (phase: Phase) = AggregatorProxy_current_phase.read();
    let (decimals) = IAggregator.decimals(contract_address=phase.aggregator);
    return (decimals,);
}

@view
func type_and_version{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}() -> (
    meta: felt
) {
    let (phase: Phase) = AggregatorProxy_current_phase.read();
    let (meta) = IAggregator.type_and_version(contract_address=phase.aggregator);
    return (meta,);
}
