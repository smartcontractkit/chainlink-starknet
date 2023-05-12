use starknet::testing::set_caller_address;
use starknet::ContractAddress;
use starknet::contract_address_const;
use starknet::class_hash::class_hash_const;
use starknet::class_hash::Felt252TryIntoClassHash;
use starknet::syscalls::deploy_syscall;

use array::ArrayTrait;
use traits::Into;
use traits::TryInto;
use option::OptionTrait;
use core::result::ResultTrait;

use chainlink::ocr2::aggregator::pow;
use chainlink::ocr2::aggregator::Aggregator;
use chainlink::tests::test_ownable::should_implement_ownable;

// TODO: aggregator tests

#[test]
#[available_gas(10000000)]
fn test_pow_2_0() {
    assert(pow(2, 0) == 0x1, 'expected 0x1');
    assert(pow(2, 1) == 0x2, 'expected 0x2');
    assert(pow(2, 2) == 0x4, 'expected 0x4');
    assert(pow(2, 3) == 0x8, 'expected 0x8');
    assert(pow(2, 4) == 0x10, 'expected 0x10');
    assert(pow(2, 5) == 0x20, 'expected 0x20');
    assert(pow(2, 6) == 0x40, 'expected 0x40');
    assert(pow(2, 7) == 0x80, 'expected 0x80');
    assert(pow(2, 8) == 0x100, 'expected 0x100');
    assert(pow(2, 9) == 0x200, 'expected 0x200');
    assert(pow(2, 10) == 0x400, 'expected 0x400');
    assert(pow(2, 11) == 0x800, 'expected 0x800');
    assert(pow(2, 12) == 0x1000, 'expected 0x1000');
    assert(pow(2, 13) == 0x2000, 'expected 0x2000');
    assert(pow(2, 14) == 0x4000, 'expected 0x4000');
    assert(pow(2, 15) == 0x8000, 'expected 0x8000');
    assert(pow(2, 16) == 0x10000, 'expected 0x10000');
    assert(pow(2, 17) == 0x20000, 'expected 0x20000');
    assert(pow(2, 18) == 0x40000, 'expected 0x40000');
    assert(pow(2, 19) == 0x80000, 'expected 0x80000');
    assert(pow(2, 20) == 0x100000, 'expected 0x100000');
    assert(pow(2, 21) == 0x200000, 'expected 0x200000');
    assert(pow(2, 22) == 0x400000, 'expected 0x400000');
    assert(pow(2, 23) == 0x800000, 'expected 0x800000');
    assert(pow(2, 24) == 0x1000000, 'expected 0x1000000');
    assert(pow(2, 25) == 0x2000000, 'expected 0x2000000');
    assert(pow(2, 26) == 0x4000000, 'expected 0x4000000');
    assert(pow(2, 27) == 0x8000000, 'expected 0x8000000');
    assert(pow(2, 28) == 0x10000000, 'expected 0x10000000');
    assert(pow(2, 29) == 0x20000000, 'expected 0x20000000');
    assert(pow(2, 30) == 0x40000000, 'expected 0x40000000');
    assert(pow(2, 31) == 0x80000000, 'expected 0x80000000');
}

fn setup() -> ContractAddress {
    let account: ContractAddress = contract_address_const::<777>();
    set_caller_address(account);
    account
}

#[test]
#[available_gas(2000000)]
fn test_ownable() {
    let account = setup();
    // Deploy aggregator
    let mut calldata = ArrayTrait::new();
    calldata.append(account.into()); // owner
    calldata.append(contract_address_const::<777>().into()); // link token
    calldata.append(0); // min_answer
    calldata.append(100); // max_answer
    calldata.append(contract_address_const::<999>().into()); // billing access controller
    calldata.append(8); // decimals
    calldata.append(123); // description
    let (aggregatorAddr, _) = deploy_syscall(
        Aggregator::TEST_CLASS_HASH.try_into().unwrap(), 0, calldata.span(), false
    ).unwrap();

    should_implement_ownable(aggregatorAddr, account);
}


#[test]
#[available_gas(2000000)]
#[should_panic(expected: ('Ownable: caller is not owner', ))]
fn test_upgrade_non_owner() {
    let sender = setup();
    Aggregator::upgrade(class_hash_const::<123>());
}
