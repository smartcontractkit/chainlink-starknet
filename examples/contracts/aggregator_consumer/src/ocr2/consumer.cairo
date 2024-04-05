#[starknet::interface]
pub trait IAggregatorConsumer<TContractState> {
    fn read_latest_round(self: @TContractState) -> chainlink::ocr2::aggregator::Round;
    fn read_ocr_address(self: @TContractState) -> starknet::ContractAddress;
    fn read_answer(self: @TContractState) -> u128;
    fn set_answer(ref self: TContractState, answer: u128);
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
        fn read_latest_round(self: @ContractState) -> Round {
            return IAggregatorDispatcher { contract_address: self._ocr_address.read() }
                .latest_round_data();
        }


        fn set_answer(ref self: ContractState, answer: u128) {
            self._answer.write(answer);
        }

        fn read_answer(self: @ContractState) -> u128 {
            return self._answer.read();
        }

        fn read_ocr_address(self: @ContractState) -> ContractAddress {
            return self._ocr_address.read();
        }
    }
}
