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

fn STATE() -> AccessController::ContractState {
    AccessController::contract_state_for_testing()
}

fn setup() -> ContractAddress {
    let account: ContractAddress = contract_address_const::<777>();
    set_caller_address(account);
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
    let (accessControllerAddr, _) = deploy_syscall(
        AccessController::TEST_CLASS_HASH.try_into().unwrap(), 0, calldata.span(), false
    )
        .unwrap();

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

    set_contract_address(owner); // required to call contract as owner

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
