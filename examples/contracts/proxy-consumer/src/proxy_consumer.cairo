#[contract]
mod ProxyConsumer {
    use zeroable::Zeroable;
    use traits::Into;
    use traits::TryInto;
    use option::OptionTrait;

    use starknet::ContractAddress;
    use starknet::StorageAccess;
    use starknet::StorageBaseAddress;
    use starknet::SyscallResult;
    use starknet::storage_read_syscall;
    use starknet::storage_write_syscall;
    use starknet::storage_address_from_base_and_offset;

    use chainlink::ocr2::aggregator::Round;

    use chainlink::ocr2::aggregator_proxy::IAggregator;
    use chainlink::ocr2::aggregator_proxy::IAggregatorDispatcher;
    use chainlink::ocr2::aggregator_proxy::IAggregatorDispatcherTrait;


    impl RoundStorageAccess of StorageAccess<Round> {
        fn read(address_domain: u32, base: StorageBaseAddress) -> SyscallResult::<Round> {
            let round_id = storage_read_syscall(
                address_domain, storage_address_from_base_and_offset(base, 0_u8)
            )?;
            let answer = storage_read_syscall(
                address_domain, storage_address_from_base_and_offset(base, 1_u8)
            )?
                .try_into()
                .unwrap();
            let block_num = storage_read_syscall(
                address_domain, storage_address_from_base_and_offset(base, 2_u8)
            )?
                .try_into()
                .unwrap();
            let started_at = storage_read_syscall(
                address_domain, storage_address_from_base_and_offset(base, 3_u8)
            )?
                .try_into()
                .unwrap();
            let updated_at = storage_read_syscall(
                address_domain, storage_address_from_base_and_offset(base, 4_u8)
            )?
                .try_into()
                .unwrap();

            Result::Ok(
                Round {
                    round_id: round_id,
                    answer: answer,
                    block_num: block_num,
                    started_at: started_at,
                    updated_at: updated_at
                }
            )
        }

        fn write(
            address_domain: u32, base: StorageBaseAddress, value: Round
        ) -> SyscallResult::<()> {
            storage_write_syscall(
                address_domain, storage_address_from_base_and_offset(base, 0_u8), value.round_id
            )?;
            storage_write_syscall(
                address_domain,
                storage_address_from_base_and_offset(base, 1_u8),
                value.answer.into()
            )?;
            storage_write_syscall(
                address_domain,
                storage_address_from_base_and_offset(base, 2_u8),
                value.block_num.into()
            )?;
            storage_write_syscall(
                address_domain,
                storage_address_from_base_and_offset(base, 3_u8),
                value.started_at.into()
            )?;
            storage_write_syscall(
                address_domain,
                storage_address_from_base_and_offset(base, 4_u8),
                value.updated_at.into()
            )
        }
    }

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
        let round = IAggregatorDispatcher {
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
