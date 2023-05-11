use starknet::ContractAddress;
use starknet::testing::set_caller_address;
use starknet::contract_address_const;
use starknet::class_hash::class_hash_const;
use starknet::syscalls::deploy_syscall;
use starknet::class_hash::Felt252TryIntoClassHash;

use array::ArrayTrait;
use traits::Into;
use traits::TryInto;
use zeroable::Zeroable;
use option::OptionTrait;
use core::result::ResultTrait;

use chainlink::libraries::simple_write_access_controller::SimpleWriteAccessController;
use chainlink::tests::test_ownable::should_implement_ownable;

fn setup() -> ContractAddress {
    let account: ContractAddress = contract_address_const::<777>();
    set_caller_address(account);
    account
}

#[test]
#[available_gas(2000000)]
fn test_ownable() {
    let account = setup();
    // Deploy simple write access controller
    let mut calldata = ArrayTrait::new();
    calldata.append(account.into()); // owner
    let (simpleWriteAccessControllerAddr, _) = deploy_syscall(
        SimpleWriteAccessController::TEST_CLASS_HASH.try_into().unwrap(), 0, calldata.span(), false
    ).unwrap();

    should_implement_ownable(simpleWriteAccessControllerAddr, account);
}

#[test]
#[available_gas(2000000)]
#[should_panic(expected: ('Ownable: caller is not owner', ))]
fn test_upgrade_not_owner() {
    let sender = setup();

    SimpleWriteAccessController::upgrade(class_hash_const::<2>());
}
