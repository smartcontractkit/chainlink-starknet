use starknet::contract_address_const;
use starknet::get_caller_address;
use starknet::ContractAddress;

use aggregator_consumer::storage::{IStorageDispatcherTrait, IStorageDispatcher};

use snforge_std::{declare, ContractClassTrait};

fn deploy_storage() -> (ContractAddress, IStorageDispatcher) {
    let mut calldata = ArrayTrait::new();
    let (contract_address, _) = declare("Storage").unwrap().deploy(@calldata).unwrap();
    (contract_address, IStorageDispatcher { contract_address: contract_address })
}

#[test]
fn test_store_and_retrieve() {
    let (_, storage) = deploy_storage();
    let desired_value: u256 = 145;
    storage.store(desired_value);

    let actual_value = storage.retrieve();

    assert(actual_value == desired_value, 'values should equal');
}

