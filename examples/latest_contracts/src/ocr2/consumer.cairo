
#[contract]
mod AggregatorConsumer {
    use starknet::ContractAddress;

    use chainlink::ocr2::aggregator::Round;

    use chainlink::ocr2::aggregator_proxy::IAggregator;
    use chainlink::ocr2::aggregator_proxy::IAggregatorDispatcher;
    use chainlink::ocr2::aggregator_proxy::IAggregatorDispatcherTrait;

    struct Storage {
        _ocr_address: ContractAddress,
    }

    #[constructor]
    fn constructor(ocr_address: ContractAddress) {
        _ocr_address::write(ocr_address);
    }

    #[view]
    fn read_latest_round() -> Round {
        IAggregatorDispatcher { 
            contract_address: _ocr_address::read() 
        }.latest_round_data()
    }

    #[view]
    fn read_decimals() -> u8 {
           IAggregatorDispatcher { 
            contract_address: _ocr_address::read() 
        }.decimals()
    }

}
