use traits::Into;

use starknet::testing::set_caller_address;
use starknet::ContractAddress;
use starknet::contract_address_const;
use starknet::class_hash::class_hash_const;
use starknet::syscalls::deploy_syscall;

use chainlink::libraries::upgradeable::Upgradeable;
use chainlink::libraries::ownable::Ownable;
use chainlink::libraries::mocks::mock_upgradeable::{
    MockUpgradeable, IMockUpgradeableDispatcher, IMockUpgradeableDispatcherTrait,
    IMockUpgradeableDispatcherImpl
};
use chainlink::libraries::mocks::mock_non_upgradeable::{
    MockNonUpgradeable, IMockNonUpgradeableDispatcher, IMockNonUpgradeableDispatcherTrait,
    IMockNonUpgradeableDispatcherImpl
};

fn STATE() -> MockUpgradeable::ContractState {
    MockUpgradeable::contract_state_for_testing()
}

// Some tests are still written in TS due to missing features in cairo-test

fn setup() -> ContractAddress {
    let account: ContractAddress = contract_address_const::<777>();
    set_caller_address(account);
    account
}

#[test]
#[available_gas(2000000)]
fn test_upgrade() {
    let sender = setup();

    // doesn't error
    Upgradeable::upgrade(class_hash_const::<1>());
}

#[test]
#[available_gas(2000000)]
fn test_upgrade_and_call() {
    let sender = setup();

    let calldata = array![];
    let (contractAddr, _) = deploy_syscall(
        MockUpgradeable::TEST_CLASS_HASH.try_into().unwrap(), 0, calldata.span(), false
    )
        .unwrap();
    let mockUpgradeable = IMockUpgradeableDispatcher { contract_address: contractAddr };
    assert(mockUpgradeable.foo() == true, 'should call foo');

    mockUpgradeable.upgrade(MockNonUpgradeable::TEST_CLASS_HASH.try_into().unwrap());

    // now, contract should be different
    let mockNonUpgradeable = IMockNonUpgradeableDispatcher { contract_address: contractAddr };
    assert(mockNonUpgradeable.bar() == true, 'should call bar');
}


#[test]
#[available_gas(2000000)]
#[should_panic(expected: ('Class hash cannot be zero',))]
fn test_upgrade_zero_hash() {
    let sender = setup();

    Upgradeable::upgrade(class_hash_const::<0>());
}
