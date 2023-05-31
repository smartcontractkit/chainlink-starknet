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

fn setup() -> ContractAddress {
    let account: ContractAddress = contract_address_const::<777>();
    set_caller_address(account);
    account
}

#[test]
#[available_gas(2000000)]
#[should_panic(expected: ('Ownable: caller is not owner', ))]
fn test_upgrade_not_owner() {
    let sender = setup();

    AccessController::upgrade(class_hash_const::<2>());
}

#[test]
#[available_gas(2000000)]
fn test_access_control() {
    let owner = setup();
    // Deploy access controller
    let mut calldata = ArrayTrait::new();
    calldata.append(owner.into()); // owner
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

#[abi]
trait IAccessController {
    fn has_access(user: ContractAddress, data: Array<felt252>) -> bool;
    fn add_access(user: ContractAddress);
    fn remove_access(user: ContractAddress);
    fn enable_access_check();
    fn disable_access_check();
}

fn should_implement_access_control(contract_addr: ContractAddress, owner: ContractAddress) {
    let contract = IAccessControllerDispatcher { contract_address: contract_addr };
    let acc2: ContractAddress = contract_address_const::<2222987765>();

    set_contract_address(owner); // required to call contract as owner

    // access check is enabled by default
    assert(!contract.has_access(acc2, ArrayTrait::new()), 'should not have access');

    // disable access check
    contract.disable_access_check();
    assert(contract.has_access(acc2, ArrayTrait::new()), 'should have access');

    // enable access check
    contract.enable_access_check();
    assert(!contract.has_access(acc2, ArrayTrait::new()), 'should not have access');

    // add_access for acc2
    contract.add_access(acc2);
    assert(contract.has_access(acc2, ArrayTrait::new()), 'should have access');

    // remove_access for acc2
    contract.remove_access(acc2);
    assert(!contract.has_access(acc2, ArrayTrait::new()), 'should not have access');
}
