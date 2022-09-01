%lang starknet

struct Round:
    member round_id : felt
    member answer : felt
    member block_num : felt
    member started_at : felt
    member updated_at : felt
end

@event
func NewTransmission(
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

@event
func AnswerUpdated(current : felt, round_id : felt, timestamp : felt):
end

@event
func NewRound(round_id : felt, started_by : felt, started_at : felt):
end

@contract_interface
namespace IAggregator:
    func latest_round_data() -> (round : Round):
    end

    func round_data(round_id : felt) -> (round : Round):
    end

    func description() -> (description : felt):
    end

    func decimals() -> (decimals : felt):
    end

    func type_and_version() -> (meta : felt):
    end
end
