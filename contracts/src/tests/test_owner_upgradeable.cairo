use traits::Into;
use zeroable::Zeroable;

use starknet::testing::set_caller_address;
use starknet::ContractAddress;
use starknet::contract_address_const;
use starknet::class_hash::class_hash_const;
use starknet::syscalls::deploy_syscall;

use openzeppelin::upgrades::interface::{
    IUpgradeable, IUpgradeableDispatcher, IUpgradeableDispatcherTrait,
};

use chainlink::libraries::upgrades::v2::owner_upgradeable::OwnerUpgradeableComponent::OwnerUpgradeableImpl;
use chainlink::libraries::upgrades::v2::owner_upgradeable::OwnerUpgradeableComponent;
use chainlink::libraries::mocks::mock_owner_upgradeable::{
    MockOwnerUpgradeable, IFoo, IFooDispatcher, IFooDispatcherTrait,
};
use chainlink::libraries::mocks::mock_non_upgradeable::{
    MockNonUpgradeable, IMockNonUpgradeableDispatcher, IMockNonUpgradeableDispatcherTrait,
    IMockNonUpgradeableDispatcherImpl,
};

use snforge_std::{
    declare, ContractClassTrait, start_cheat_caller_address_global,
    stop_cheat_caller_address_global, DeclareResultTrait,
};

fn setup() -> ContractAddress {
    let account: ContractAddress = contract_address_const::<777>();
    start_cheat_caller_address_global(account);
    account
}

fn STATE() -> MockOwnerUpgradeable::ContractState {
    MockOwnerUpgradeable::contract_state_for_testing()
}

#[test]
fn test_upgrade_and_call() {
    let account = setup();

    let calldata = array![account.into()];

    let (contractAddr, _) = declare("MockOwnerUpgradeable")
        .unwrap()
        .contract_class()
        .deploy(@calldata)
        .unwrap();

    let mockUpgradeable = IFooDispatcher { contract_address: contractAddr };
    assert(mockUpgradeable.foo() == true, 'should call foo');

    let contract = declare("MockNonUpgradeable").unwrap().contract_class();

    let mockUpgradeable = IUpgradeableDispatcher { contract_address: contractAddr };

    mockUpgradeable.upgrade(*(contract.class_hash));

    // now, contract should be different
    let mockNonUpgradeable = IMockNonUpgradeableDispatcher { contract_address: contractAddr };
    assert(mockNonUpgradeable.bar() == true, 'should call bar');
}

#[test]
#[should_panic(expected: ('Caller is not the owner',))]
fn test_upgrade_non_owner() {
    let _ = setup();

    let mut state = STATE();

    OwnerUpgradeableImpl::upgrade(ref state, class_hash_const::<0>());
}

#[test]
#[should_panic(expected: ('Class hash cannot be zero',))]
fn test_upgrade_zero() {
    let account = setup();

    let calldata = array![account.into()];

    let (contractAddr, _) = declare("MockOwnerUpgradeable")
        .unwrap()
        .contract_class()
        .deploy(@calldata)
        .unwrap();

    let mockUpgradeable = IFooDispatcher { contract_address: contractAddr };
    assert(mockUpgradeable.foo() == true, 'should call foo');

    let mockUpgradeable = IUpgradeableDispatcher { contract_address: contractAddr };

    mockUpgradeable.upgrade(Zeroable::zero());
}

