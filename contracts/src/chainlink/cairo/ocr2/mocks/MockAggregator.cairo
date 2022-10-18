%lang starknet

from starkware.cairo.common.alloc import alloc
from starkware.cairo.common.cairo_builtins import HashBuiltin
from chainlink.cairo.ocr2.IAggregator import NewTransmission, Round

struct Transmission:
    member answer : felt
    member block_num : felt
    member observation_timestamp : felt
    member transmission_timestamp : felt
end

@storage_var
func MockAggregator_transmissions(round_id : felt) -> (transmission : Transmission):
end

@storage_var
func MockAggregator_latest_aggregator_round_id() -> (round_id : felt):
end

@storage_var
func MockAggregator_decimals() -> (decimals : felt):
end

@constructor
func constructor{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    decimals : felt
):
    MockAggregator_decimals.write(decimals)
    return ()
end

@external
func set_latest_round_data{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    answer : felt, block_num : felt, observation_timestamp : felt, transmission_timestamp : felt
):
    alloc_locals
    let (prev_round_id) = MockAggregator_latest_aggregator_round_id.read()
    let round_id = prev_round_id + 1
    MockAggregator_latest_aggregator_round_id.write(round_id)
    MockAggregator_transmissions.write(
        round_id,
        Transmission(
        answer=answer,
        block_num=block_num,
        observation_timestamp=observation_timestamp,
        transmission_timestamp=transmission_timestamp,
        ),
    )

    let (observations : felt*) = alloc()
    assert observations[0] = 2
    assert observations[1] = 3
    NewTransmission.emit(
        round_id=round_id,
        answer=answer,
        transmitter=12,
        observation_timestamp=observation_timestamp,
        observers=3,
        observations_len=2,
        observations=observations,
        juels_per_fee_coin=18,
        gas_price=1,
        config_digest=34,
        epoch_and_round=20,
        reimbursement=100,
    )
    return ()
end

@view
func latest_round_data{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    round : Round
):
    alloc_locals
    let (latest_round_id) = MockAggregator_latest_aggregator_round_id.read()
    let (transmission : Transmission) = MockAggregator_transmissions.read(latest_round_id)

    let round = Round(
        round_id=latest_round_id,
        answer=transmission.answer,
        block_num=transmission.block_num,
        started_at=transmission.observation_timestamp,
        updated_at=transmission.transmission_timestamp,
    )
    return (round)
end

@view
func decimals{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    decimals : felt
):
    let (decimals) = MockAggregator_decimals.read()
    return (decimals)
end
