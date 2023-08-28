#[starknet::contract]
mod ProxyConsumer {
    use zeroable::Zeroable;
    use traits::Into;
    use traits::TryInto;
    use option::OptionTrait;

    use starknet::ContractAddress;
    use starknet::StorageBaseAddress;
    use starknet::SyscallResult;
    use starknet::storage_read_syscall;
    use starknet::storage_write_syscall;
    use starknet::storage_address_from_base_and_offset;

    use chainlink::ocr2::aggregator::Round;

    use chainlink::ocr2::aggregator_proxy::IAggregator;
    use chainlink::ocr2::aggregator_proxy::IAggregatorDispatcher;
    use chainlink::ocr2::aggregator_proxy::IAggregatorDispatcherTrait;


    #[storage]
    struct Storage {
        _proxy_address: ContractAddress,
        _feed_data: Round,
    }

    #[constructor]
    fn constructor(ref self: ContractState, proxy_address: ContractAddress) {
        assert(!proxy_address.is_zero(), 'proxy address 0');
        self._proxy_address.write(proxy_address);
        get_latest_round_data(ref self);
    }

    #[external(v0)]
    fn get_latest_round_data(ref self: ContractState) -> Round {
        let round = IAggregatorDispatcher { contract_address: self._proxy_address.read() }
            .latest_round_data();
        self._feed_data.write(round);
        round
    }

    #[external(v0)]
    fn get_stored_round(self: @ContractState) -> Round {
        self._feed_data.read()
    }

    #[external(v0)]
    fn get_stored_feed_address(self: @ContractState) -> ContractAddress {
        self._proxy_address.read()
    }
}
