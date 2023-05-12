use starknet::ContractAddress;
use starknet::testing::set_caller_address;
use starknet::contract_address_const;
use starknet::class_hash::class_hash_const;
use starknet::class_hash::Felt252TryIntoClassHash;
use starknet::syscalls::deploy_syscall;

use array::ArrayTrait;
use traits::Into;
use traits::TryInto;
use zeroable::Zeroable;
use option::OptionTrait;
use core::result::ResultTrait;

use chainlink::token::link_token::LinkToken;
use chainlink::tests::test_ownable::should_implement_ownable;

// only tests link token specific functionality 
// erc20 and erc677 functionality is already tested elsewhere

fn setup() -> ContractAddress {
    let account: ContractAddress = contract_address_const::<1>();
    // Set account as default caller
    set_caller_address(account);
    account
}

#[test]
#[available_gas(2000000)]
fn test_ownable() {
    let account = setup();
    // Deploy LINK token
    let mut calldata = ArrayTrait::new();
    calldata.append(class_hash_const::<123>().into()); // minter
    calldata.append(account.into()); // owner
    let (linkAddr, _) = deploy_syscall(
        LinkToken::TEST_CLASS_HASH.try_into().unwrap(), 0, calldata.span(), false
    ).unwrap();

    should_implement_ownable(linkAddr, account);
}

#[test]
#[available_gas(2000000)]
#[should_panic(expected: ('minter is 0', ))]
fn test_constructor_zero_address() {
    let sender = setup();

    LinkToken::constructor(Zeroable::zero(), sender);
}

#[test]
#[available_gas(2000000)]
fn test_constructor() {
    let sender = setup();
    LinkToken::constructor(sender, sender);

    assert(LinkToken::minter() == sender, 'minter valid');
    assert(LinkToken::name() == 'ChainLink Token', 'name valid');
    assert(LinkToken::symbol() == 'LINK', 'symbol valid');
}

#[test]
#[available_gas(2000000)]
fn test_permissioned_mint_from_minter() {
    let sender = setup();
    LinkToken::constructor(sender, sender);
    let to = contract_address_const::<908>();

    let zero: felt252 = 0;
    assert(LinkToken::balance_of(sender) == zero.into(), 'zero balance');
    assert(LinkToken::balance_of(to) == zero.into(), 'zero balance');

    let amount: felt252 = 3000;
    LinkToken::permissionedMint(to, amount.into());

    assert(LinkToken::balance_of(sender) == zero.into(), 'zero balance');
    assert(LinkToken::balance_of(to) == amount.into(), 'expect balance');
}

#[test]
#[available_gas(2000000)]
#[should_panic(expected: ('only minter', ))]
fn test_permissioned_mint_from_nonminter() {
    let sender = setup();
    let minter = contract_address_const::<111>();
    LinkToken::constructor(minter, sender);
    let to = contract_address_const::<908>();

    let amount: felt252 = 3000;
    LinkToken::permissionedMint(to, amount.into());
}

#[test]
#[available_gas(2000000)]
#[should_panic(expected: ('u256_sub Overflow', ))]
fn test_permissioned_burn_from_minter() {
    let zero = 0;
    let sender = setup();
    LinkToken::constructor(sender, sender);
    let to = contract_address_const::<908>();

    let amount: felt252 = 3000;
    LinkToken::permissionedMint(to, amount.into());
    assert(LinkToken::balance_of(to) == amount.into(), 'expect balance');

    // burn some
    let burn_amount: felt252 = 2000;
    let remaining_amount: felt252 = amount - burn_amount;
    LinkToken::permissionedBurn(to, burn_amount.into());
    assert(LinkToken::balance_of(to) == remaining_amount.into(), 'remaining balance');

    // burn remaining
    LinkToken::permissionedBurn(to, remaining_amount.into());
    assert(LinkToken::balance_of(to) == zero.into(), 'no balance');

    // burn too much
    LinkToken::permissionedBurn(to, amount.into());
}


#[test]
#[available_gas(2000000)]
#[should_panic(expected: ('only minter', ))]
fn test_permissioned_burn_from_nonminter() {
    let sender = setup();
    let minter = contract_address_const::<111>();
    LinkToken::constructor(minter, sender);
    let to = contract_address_const::<908>();

    let amount: felt252 = 3000;
    LinkToken::permissionedBurn(to, amount.into());
}

#[test]
#[available_gas(2000000)]
#[should_panic(expected: ('Ownable: caller is not owner', ))]
fn test_upgrade_non_owner() {
    let sender = setup();
    LinkToken::constructor(sender, contract_address_const::<111>());

    LinkToken::upgrade(class_hash_const::<123>());
}

