%lang starknet

from starkware.cairo.common.alloc import alloc
from starkware.cairo.common.cairo_builtins import HashBuiltin
from contracts.ocr2.interfaces.IAggregator import IAggregator, Round

struct Transmission:
    member answer : felt
    member block_num : felt
    member observation_timestamp : felt
    member transmission_timestamp : felt
end

@storage_var
func transmissions_(round_id : felt) -> (transmission : Transmission):
end

@storage_var
func latest_aggregator_round_id_() -> (round_id : felt):
end

@storage_var
func decimals_() -> (decimals : felt):
end

@storage_var
func answer_() -> (answer : felt):
end

@event
func new_transmission(
    round_id : felt,
    answer : felt,
    transmitter : felt,
    observation_timestamp : felt,
    observers : felt,
    observations_len : felt,
    observations : felt*,
    juels_per_fee_coin : felt,
    config_digest : felt,
    epoch_and_round : felt,
    reimbursement : felt,
):
end

@constructor
func constructor{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    decimals : felt
):
    decimals_.write(decimals)
    return ()
end

@external
func set_latest_round_data{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    answer : felt, block_num : felt, observation_timestamp : felt, transmission_timestamp : felt
):
    alloc_locals
    let (prev_round_id) = latest_aggregator_round_id_.read()
    let round_id = prev_round_id + 1
    latest_aggregator_round_id_.write(round_id)
    transmissions_.write(
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
    new_transmission.emit(
        round_id=round_id,
        answer=answer,
        transmitter=12,
        observation_timestamp=observation_timestamp,
        observers=3,
        observations_len=2,
        observations=observations,
        juels_per_fee_coin=18,
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
    let (latest_round_id) = latest_aggregator_round_id_.read()
    let (transmission : Transmission) = transmissions_.read(latest_round_id)

    let round = Round(
        round_id=latest_round_id,
        answer=transmission.answer,
        block_num=transmission.block_num,
        started_at=transmission.observation_timestamp,
        updated_at=transmission.transmission_timestamp,
    )
    let (observations : felt*) = alloc()
    assert observations[0] = 2
    assert observations[1] = 3
    new_transmission.emit(
        round_id=latest_round_id,
        answer=transmission.answer,
        transmitter=12,
        observation_timestamp=transmission.observation_timestamp,
        observers=3,
        observations_len=2,
        observations=observations,
        juels_per_fee_coin=18,
        config_digest=34,
        epoch_and_round=20,
        reimbursement=100,
    )
    return (round)
end

@view
func decimals{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    decimals : felt
):
    let (decimals) = decimals_.read()
    return (decimals)
end
