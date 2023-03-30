%lang starknet

struct Round {
    round_id: felt,
    answer: felt,
    block_num: felt,
    started_at: felt,
    updated_at: felt,
}

@event
func NewTransmission(
    round_id: felt,
    answer: felt,
    transmitter: felt,
    observation_timestamp: felt,
    observers: felt,
    observations_len: felt,
    observations: felt*,
    juels_per_fee_coin: felt,
    gas_price: felt,
    config_digest: felt,
    epoch_and_round: felt,
    reimbursement: felt,
) {
}

@event
func AnswerUpdated(current: felt, round_id: felt, timestamp: felt) {
}

@event
func NewRound(round_id: felt, started_by: felt, started_at: felt) {
}

@contract_interface
namespace IAggregator {
    func latest_round_data() -> (round: Round) {
    }

    func round_data(round_id: felt) -> (round: Round) {
    }

    func description() -> (description: felt) {
    }

    func decimals() -> (decimals: felt) {
    }

    func type_and_version() -> (meta: felt) {
    }
}
