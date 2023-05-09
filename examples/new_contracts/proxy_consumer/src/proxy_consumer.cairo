
#[contract]
mod ProxyConsumer {
    use starknet::ContractAddress;
    use starknet::contract_address::ContractAddressZeroable;
    use chainlink::ocr2::aggregator::Round;

    use chainlink::ocr2::aggregator_proxy::IAggregator;
    use chainlink::ocr2::aggregator_proxy::IAggregatorDispatcher;
    use chainlink::ocr2::aggregator_proxy::IAggregatorDispatcherTrait;

    struct Storage {
        _proxy_address: ContractAddress,
        _feed_data: Round,
    }

    #[constructor]
    fn constructor(proxy_address: ContractAddress) {
        assert(!proxy_address.is_zero(), 'proxy address 0');
        _proxy_address::write(proxy_address);
        get_latest_round_data();
    } 

    #[external]
    fn get_latest_round_data() -> Round {
        let round = IAggregatorDispatcher{
            contract_address: _proxy_address::read()
        }.latest_round_data();
        _feed_data::write(round);
        round
    }

    #[view]
    fn get_stored_round() -> Round {
        _feed_data::read()
    }

    #[view]
    fn get_stored_feed_address() -> ContractAddress {
        _proxy_address::read()
    }
}
