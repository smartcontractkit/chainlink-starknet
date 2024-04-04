#[starknet::interface]
pub trait IAggregatorConsumer<TContractState> {
    fn read_answer(self: @TContractState) -> u128;
    fn set_answer(ref self: TContractState);
}

#[starknet::contract]
mod AggregatorConsumer {
    use starknet::ContractAddress;
    use traits::Into;

    use chainlink::ocr2::aggregator::Round;

    use chainlink::ocr2::aggregator_proxy::IAggregator;
    use chainlink::ocr2::aggregator_proxy::IAggregatorDispatcher;
    use chainlink::ocr2::aggregator_proxy::IAggregatorDispatcherTrait;

    #[storage]
    struct Storage {
        _ocr_address: ContractAddress,
        _answer: u128,
    }

    #[constructor]
    fn constructor(ref self: ContractState, ocr_address: ContractAddress) {
        self._ocr_address.write(ocr_address);
        self._answer.write(0);
    }

    #[abi(embed_v0)]
    impl AggregatorConsumerImpl of super::IAggregatorConsumer<ContractState> {
        fn set_answer(ref self: ContractState) {
            let round = IAggregatorDispatcher { contract_address: self._ocr_address.read() }
                .latest_round_data();
            self._answer.write(round.answer);
        }

        fn read_answer(self: @ContractState) -> u128 {
            return self._answer.read();
        }
    }
}
