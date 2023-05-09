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

use chainlink::libraries::simple_read_access_controller::SimpleReadAccessController;
use chainlink::tests::test_ownable::should_behave_like_ownable_contract;

fn setup() -> ContractAddress {
    let account: ContractAddress = contract_address_const::<777>();
    set_caller_address(account);
    account
}

#[ignore]
#[test]
#[available_gas(2000000)]
fn test_ownable() {
    let account = setup();
    // TODO: controller currently does not expose ownable functions,
    // refactor to separate lib from deployable contract
    let mut calldata = ArrayTrait::new();
    calldata.append(account.into()); // owner
    let (simpleReadAccessControllerAddr, _) = deploy_syscall(
        SimpleReadAccessController::TEST_CLASS_HASH.try_into().unwrap(), 0, calldata.span(), false
    ).unwrap();

    should_behave_like_ownable_contract(simpleReadAccessControllerAddr, account);
}

#[test]
#[available_gas(2000000)]
#[should_panic(expected: ('Ownable: caller is not owner', ))]
fn test_upgrade_not_owner() {
    let sender = setup();

    SimpleReadAccessController::upgrade(class_hash_const::<2>());
}
