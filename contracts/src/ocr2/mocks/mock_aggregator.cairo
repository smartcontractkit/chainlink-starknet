#[contract]
mod MockAggregator {
    use array::ArrayTrait;
    use starknet::contract_address_const;
     use traits::Into;

    use chainlink::ocr2::aggregator::Aggregator::Transmission;
    use chainlink::ocr2::aggregator::Aggregator::TransmissionStorageAccess;
    use chainlink::ocr2::aggregator::Aggregator::NewTransmission;
    use chainlink::ocr2::aggregator::Round;

    struct Storage {
        _transmissions: LegacyMap<u128, Transmission>,
        _latest_aggregator_round_id: u128,
        _decimals: u8
    }

    #[constructor]
    fn constructor(decimals: u8) {
        _decimals::write(decimals);
    }

    #[external]
    fn set_latest_round_data(answer: u128, block_num: u64, observation_timestamp: u64, transmission_timestamp: u64) {
        let new_round_id = _latest_aggregator_round_id::read() + 1_u128;
        _transmissions::write(
            new_round_id,
            Transmission{
                answer: answer,
                block_num: block_num,
                observation_timestamp: observation_timestamp,
                transmission_timestamp: transmission_timestamp
            }
        );

        let mut observations = ArrayTrait::new();
        observations.append(2_u128);
        observations.append(3_u128);

        NewTransmission(
            new_round_id,
            answer,
            contract_address_const::<42>(),
            observation_timestamp,
            3,
            observations,
            18_u128,
            1_u128,
            777,
            20_u64,
            100_u128
        );
    }

    #[view]
    fn latest_round_data() -> Round {
        let latest_round_id = _latest_aggregator_round_id::read();
        let transmission = _transmissions::read(latest_round_id);

        Round{
            round_id: latest_round_id.into(),
            answer: transmission.answer,
            block_num: transmission.block_num,
            started_at: transmission.observation_timestamp,
            updated_at: transmission.transmission_timestamp
        }
    }

    #[view]
    fn decimals() -> u8 {
        _decimals::read()
    }

}
