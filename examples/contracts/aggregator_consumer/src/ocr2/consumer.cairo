#[starknet::interface]
pub trait IAggregatorConsumer<TContractState> {
    fn read_latest_round(self: @TContractState) -> chainlink::ocr2::aggregator::Round;
    fn read_decimals(self: @TContractState) -> u8;
}

#[starknet::contract]
mod AggregatorConsumer {
    use starknet::ContractAddress;

    use chainlink::ocr2::aggregator::Round;

    use chainlink::ocr2::aggregator_proxy::IAggregator;
    use chainlink::ocr2::aggregator_proxy::IAggregatorDispatcher;
    use chainlink::ocr2::aggregator_proxy::IAggregatorDispatcherTrait;

    #[storage]
    struct Storage {
        _ocr_address: ContractAddress,
    }

    #[constructor]
    fn constructor(ref self: ContractState, ocr_address: ContractAddress) {
        self._ocr_address.write(ocr_address);
    }

    #[abi(embed_v0)]
    impl AggregatorConsumerImpl of super::IAggregatorConsumer<ContractState> {
        fn read_latest_round(self: @ContractState) -> Round {
            IAggregatorDispatcher { contract_address: self._ocr_address.read() }.latest_round_data()
        }

        fn read_decimals(self: @ContractState) -> u8 {
            IAggregatorDispatcher { contract_address: self._ocr_address.read() }.decimals()
        }
    }
}
