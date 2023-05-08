use traits::Into;

use starknet::testing::set_caller_address;
use starknet::ContractAddress;
use starknet::contract_address_const;
use starknet::class_hash::class_hash_const;

use chainlink::libraries::upgradeable::Upgradeable;
use chainlink::libraries::ownable::Ownable;


fn setup() -> ContractAddress {
    let account: ContractAddress = contract_address_const::<777>();
    set_caller_address(account);
    account
}

// replace_class_syscall() not yet supported in tests
#[test]
#[ignore]
#[available_gas(2000000)]
fn test_upgrade() {
    let sender = setup();

    Upgradeable::upgrade(class_hash_const::<1>());
}

#[test]
#[available_gas(2000000)]
#[should_panic(expected: ('Class hash cannot be zero', ))]
fn test_upgrade_zero_hash() {
    let sender = setup();

    Upgradeable::upgrade(class_hash_const::<0>());
}
