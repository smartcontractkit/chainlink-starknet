%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin
from starkware.cairo.common.math import split_felt, assert_not_zero
from starkware.starknet.common.syscalls import get_caller_address

from chainlink.cairo.ocr2.IAggregator import IAggregator, Round

from chainlink.cairo.access.SimpleReadAccessController.library import SimpleReadAccessController
from chainlink.cairo.access.ownable import Ownable

struct Phase:
    member id : felt
    member aggregator : felt
end

@storage_var
func Proxy_current_phase() -> (phase : Phase):
end

@storage_var
func Proxy_proposed_aggregator() -> (address : felt):
end

@storage_var
func Proxy_phases(id : felt) -> (address : felt):
end

const SHIFT = 2 ** 128
const MAX_ID = SHIFT - 1

@event
func AggregatorProposed(current : felt, proposed : felt):
end

@event
func AggregatorConfirmed(previous : felt, latest : felt):
end

@constructor
func constructor{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    owner : felt, address : felt
):
    SimpleReadAccessController.initialize(owner)  # This also calls Ownable.initializer
    set_aggregator(address)
    return ()
end

func set_aggregator{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    address : felt
):
    let (current_phase : Phase) = Proxy_current_phase.read()
    let id = current_phase.id + 1
    Proxy_current_phase.write(Phase(id=id, aggregator=address))
    Proxy_phases.write(id, address)
    return ()
end

@external
func propose_aggregator{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    address : felt
):
    Ownable.assert_only_owner()

    Proxy_proposed_aggregator.write(address)

    # emit event
    let (phase : Phase) = Proxy_current_phase.read()
    AggregatorProposed.emit(current=phase.aggregator, proposed=address)
    return ()
end

@external
func confirm_aggregator{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    address : felt
):
    Ownable.assert_only_owner()

    let (phase : Phase) = Proxy_current_phase.read()
    let previous = phase.aggregator

    let (proposed_aggregator) = Proxy_proposed_aggregator.read()
    assert proposed_aggregator = address
    Proxy_proposed_aggregator.write(0)
    set_aggregator(proposed_aggregator)

    # emit event
    AggregatorConfirmed.emit(previous=previous, latest=address)
    return ()
end

# Read access helper
func require_access{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}():
    let (address) = get_caller_address()
    SimpleReadAccessController.check_access(address)

    return ()
end

@view
func latest_round_data{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    round : Round
):
    require_access()
    let (phase : Phase) = Proxy_current_phase.read()
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
    require_access()
    let (phase_id, round_id) = split_felt(round_id)
    let (address) = Proxy_phases.read(phase_id)
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
    require_access()
    let (aggregator) = Proxy_proposed_aggregator.read()
    let (round : Round) = IAggregator.latest_round_data(contract_address=aggregator)
    return (round)
end

@view
func proposed_round_data{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    round_id : felt
) -> (round : Round):
    require_access()
    let (aggregator) = Proxy_proposed_aggregator.read()
    let (round : Round) = IAggregator.round_data(contract_address=aggregator, round_id=round_id)
    return (round)
end

@view
func aggregator{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    aggregator : felt
):
    require_access()
    let (phase : Phase) = Proxy_current_phase.read()
    return (phase.aggregator)
end

@view
func phase_id{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    phase_id : felt
):
    require_access()
    let (phase : Phase) = Proxy_current_phase.read()
    return (phase.id)
end

@view
func description{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    description : felt
):
    require_access()
    let (phase : Phase) = Proxy_current_phase.read()
    let (description) = IAggregator.description(contract_address=phase.aggregator)
    return (description)
end

@view
func decimals{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    decimals : felt
):
    require_access()
    let (phase : Phase) = Proxy_current_phase.read()
    let (decimals) = IAggregator.decimals(contract_address=phase.aggregator)
    return (decimals)
end

@view
func type_and_version{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    meta : felt
):
    let (phase : Phase) = Proxy_current_phase.read()
    let (meta) = IAggregator.type_and_version(contract_address=phase.aggregator)
    return (meta)
end
