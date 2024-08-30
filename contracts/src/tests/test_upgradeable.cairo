use traits::Into;

use starknet::testing::set_caller_address;
use starknet::ContractAddress;
use starknet::contract_address_const;
use starknet::class_hash::class_hash_const;
use starknet::syscalls::deploy_syscall;

use chainlink::libraries::upgradeable::Upgradeable;
use chainlink::libraries::mocks::mock_upgradeable::{
    MockUpgradeable, IMockUpgradeableDispatcher, IMockUpgradeableDispatcherTrait,
    IMockUpgradeableDispatcherImpl
};
use chainlink::libraries::mocks::mock_non_upgradeable::{
    MockNonUpgradeable, IMockNonUpgradeableDispatcher, IMockNonUpgradeableDispatcherTrait,
    IMockNonUpgradeableDispatcherImpl
};

use snforge_std::{
    declare, ContractClassTrait, start_cheat_caller_address_global, stop_cheat_caller_address_global
};


fn setup() -> ContractAddress {
    let account: ContractAddress = contract_address_const::<777>();
    start_cheat_caller_address_global(account);
    account
}

#[test]
fn test_upgrade_and_call() {
    let _ = setup();

    let calldata = array![];

    let (contractAddr, _) = declare("MockUpgradeable").unwrap().deploy(@calldata).unwrap();

    let mockUpgradeable = IMockUpgradeableDispatcher { contract_address: contractAddr };
    assert(mockUpgradeable.foo() == true, 'should call foo');

    let contract_class = declare("MockNonUpgradeable").unwrap();

    mockUpgradeable.upgrade(contract_class.class_hash);

    // now, contract should be different
    let mockNonUpgradeable = IMockNonUpgradeableDispatcher { contract_address: contractAddr };
    assert(mockNonUpgradeable.bar() == true, 'should call bar');
}


#[test]
#[should_panic(expected: ('Class hash cannot be zero',))]
fn test_upgrade_zero_hash() {
    let _ = setup();

    Upgradeable::upgrade(class_hash_const::<0>());
}
