use starknet::ContractAddress;
use starknet::testing::set_caller_address;
use starknet::testing::set_contract_address;
use starknet::contract_address_const;
use starknet::class_hash::class_hash_const;
use starknet::class_hash::Felt252TryIntoClassHash;
use starknet::syscalls::deploy_syscall;

use array::ArrayTrait;
use traits::Into;
use traits::TryInto;
use option::OptionTrait;
use core::result::ResultTrait;

use chainlink::access_control::access_controller::AccessController;
use chainlink::access_control::access_controller::AccessController::UpgradeableImpl;

use chainlink::libraries::access_control::{
    IAccessController, IAccessControllerDispatcher, IAccessControllerDispatcherTrait
};

use snforge_std::{
    declare, ContractClassTrait, start_cheat_caller_address_global, stop_cheat_caller_address_global
};


fn STATE() -> AccessController::ContractState {
    AccessController::contract_state_for_testing()
}

fn setup() -> ContractAddress {
    let account: ContractAddress = contract_address_const::<777>();
    start_cheat_caller_address_global(account);
    account
}

#[test]
#[should_panic(expected: ('Caller is not the owner',))]
fn test_upgrade_not_owner() {
    let _ = setup();
    let mut state = STATE();

    UpgradeableImpl::upgrade(ref state, class_hash_const::<2>());
}

#[test]
fn test_access_control() {
    let owner = setup();
    // Deploy access controller
    let calldata = array![owner.into(), // owner
    ];

    let contract = declare("AccessController").unwrap();

    let (accessControllerAddr, _) = contract.deploy(@calldata).unwrap();

    should_implement_access_control(accessControllerAddr, owner);
}

//
// Tests for contracts that inherit AccessControl.
// Write functions are assumed to be protected by Ownable::only_owner,
// but this test does not check for that.
//

fn should_implement_access_control(contract_addr: ContractAddress, owner: ContractAddress) {
    let contract = IAccessControllerDispatcher { contract_address: contract_addr };
    let acc2: ContractAddress = contract_address_const::<2222987765>();

    start_cheat_caller_address_global(owner);

    // access check is enabled by default
    assert(!contract.has_access(acc2, array![]), 'should not have access');

    // disable access check
    contract.disable_access_check();
    assert(contract.has_access(acc2, array![]), 'should have access');

    // enable access check
    contract.enable_access_check();
    assert(!contract.has_access(acc2, array![]), 'should not have access');

    // add_access for acc2
    contract.add_access(acc2);
    assert(contract.has_access(acc2, array![]), 'should have access');

    // remove_access for acc2
    contract.remove_access(acc2);
    assert(!contract.has_access(acc2, array![]), 'should not have access');
}
