use starknet::ContractAddress;
use starknet::testing::set_caller_address;
use starknet::contract_address_const;
use starknet::class_hash::class_hash_const;

use chainlink::access_control::simple_read_access_controller::SimpleReadAccessController;

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

    SimpleReadAccessController::upgrade(class_hash_const::<2>());
}
