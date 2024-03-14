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
use chainlink::token::link_token::LinkToken::{MintableToken, UpgradeableImpl};
use openzeppelin::token::erc20::ERC20Component::{ERC20Impl, ERC20MetadataImpl};
use chainlink::tests::test_ownable::should_implement_ownable;

// only tests link token specific functionality 
// erc20 and erc677 functionality is already tested elsewhere

fn STATE() -> LinkToken::ContractState {
    LinkToken::contract_state_for_testing()
}

fn setup() -> ContractAddress {
    let account: ContractAddress = contract_address_const::<1>();
    // Set account as default caller
    set_caller_address(account);
    account
}

#[test]
fn test_ownable() {
    let account = setup();
    // Deploy LINK token
    let mut calldata = ArrayTrait::new();
    calldata.append(class_hash_const::<123>().into()); // minter
    calldata.append(account.into()); // owner
    let (linkAddr, _) = deploy_syscall(
        LinkToken::TEST_CLASS_HASH.try_into().unwrap(), 0, calldata.span(), false
    )
        .unwrap();

    should_implement_ownable(linkAddr, account);
}

#[test]
#[should_panic(expected: ('minter is 0',))]
fn test_constructor_zero_address() {
    let sender = setup();
    let mut state = STATE();

    LinkToken::constructor(ref state, Zeroable::zero(), sender);
}

#[test]
fn test_constructor() {
    let sender = setup();
    let mut state = STATE();
    LinkToken::constructor(ref state, sender, sender);

    assert(LinkToken::minter(@state) == sender, 'minter valid');
    assert(state.erc20.name() == "ChainLink Token", 'name valid');
    assert(state.erc20.symbol() == "LINK", 'symbol valid');
}

#[test]
fn test_permissioned_mint_from_minter() {
    let sender = setup();
    let mut state = STATE();
    LinkToken::constructor(ref state, sender, sender);
    let to = contract_address_const::<908>();

    let zero: felt252 = 0;
    assert(ERC20Impl::balance_of(@state, sender) == zero.into(), 'zero balance');
    assert(ERC20Impl::balance_of(@state, to) == zero.into(), 'zero balance');

    let amount: felt252 = 3000;
    MintableToken::permissioned_mint(ref state, to, amount.into());

    assert(ERC20Impl::balance_of(@state, sender) == zero.into(), 'zero balance');
    assert(ERC20Impl::balance_of(@state, to) == amount.into(), 'expect balance');
}

#[test]
#[should_panic(expected: ('only minter',))]
fn test_permissioned_mint_from_nonminter() {
    let sender = setup();
    let mut state = STATE();
    let minter = contract_address_const::<111>();
    LinkToken::constructor(ref state, minter, sender);
    let to = contract_address_const::<908>();

    let amount: felt252 = 3000;
    MintableToken::permissioned_mint(ref state, to, amount.into());
}

#[test]
#[should_panic(expected: ('u256_sub Overflow',))]
fn test_permissioned_burn_from_minter() {
    let zero = 0;
    let sender = setup();
    let mut state = STATE();
    LinkToken::constructor(ref state, sender, sender);
    let to = contract_address_const::<908>();

    let amount: felt252 = 3000;
    MintableToken::permissioned_mint(ref state, to, amount.into());
    assert(ERC20Impl::balance_of(@state, to) == amount.into(), 'expect balance');

    // burn some
    let burn_amount: felt252 = 2000;
    let remaining_amount: felt252 = amount - burn_amount;
    MintableToken::permissioned_burn(ref state, to, burn_amount.into());
    assert(ERC20Impl::balance_of(@state, to) == remaining_amount.into(), 'remaining balance');

    // burn remaining
    MintableToken::permissioned_burn(ref state, to, remaining_amount.into());
    assert(ERC20Impl::balance_of(@state, to) == zero.into(), 'no balance');

    // burn too much
    MintableToken::permissioned_burn(ref state, to, amount.into());
}


#[test]
#[should_panic(expected: ('only minter',))]
fn test_permissioned_burn_from_nonminter() {
    let sender = setup();
    let mut state = STATE();
    let minter = contract_address_const::<111>();
    LinkToken::constructor(ref state, minter, sender);
    let to = contract_address_const::<908>();

    let amount: felt252 = 3000;
    MintableToken::permissioned_burn(ref state, to, amount.into());
}

#[test]
#[should_panic(expected: ('Caller is not the owner',))]
fn test_upgrade_non_owner() {
    let sender = setup();
    let mut state = STATE();
    LinkToken::constructor(ref state, sender, contract_address_const::<111>());

    UpgradeableImpl::upgrade(ref state, class_hash_const::<123>());
}

