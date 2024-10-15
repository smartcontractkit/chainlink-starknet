use starknet::{
    syscalls::deploy_syscall, ContractAddress, testing::set_caller_address, contract_address_const,
    class_hash::{class_hash_const, Felt252TryIntoClassHash}
};

use array::ArrayTrait;
use traits::{Into, TryInto};
use zeroable::Zeroable;
use option::OptionTrait;
use core::result::ResultTrait;

use chainlink::token::v2::link_token::{
    LinkToken, LinkToken::{MintableToken, UpgradeableImpl, Minter}
};
use openzeppelin::token::erc20::ERC20Component::{ERC20Impl, ERC20MetadataImpl};
use chainlink::tests::test_ownable::should_implement_ownable;

use snforge_std::{
    declare, ContractClassTrait, start_cheat_caller_address_global, stop_cheat_caller_address_global
};


// only tests link token specific functionality 
// erc20 and erc677 functionality is already tested elsewhere

fn STATE() -> LinkToken::ContractState {
    LinkToken::contract_state_for_testing()
}

fn setup() -> ContractAddress {
    let account: ContractAddress = contract_address_const::<1>();
    // Set account as default caller
    start_cheat_caller_address_global(account);
    account
}

fn link_deploy_args(minter: ContractAddress, owner: ContractAddress) -> Array<felt252> {
    let mut calldata = ArrayTrait::new();
    let _name_ignore: felt252 = 0;
    let _symbol_ignore: felt252 = 0;
    let _decimals_ignore: u8 = 0;
    let _initial_supply_ignore: u256 = 0;
    let _initial_recipient_ignore: ContractAddress = Zeroable::zero();
    let _upgrade_delay_ignore: u64 = 0;
    Serde::serialize(@_name_ignore, ref calldata);
    Serde::serialize(@_symbol_ignore, ref calldata);
    Serde::serialize(@_decimals_ignore, ref calldata);
    Serde::serialize(@_initial_supply_ignore, ref calldata);
    Serde::serialize(@_initial_recipient_ignore, ref calldata);
    Serde::serialize(@minter, ref calldata);
    Serde::serialize(@owner, ref calldata);
    Serde::serialize(@_upgrade_delay_ignore, ref calldata);

    calldata
}

#[test]
fn test_ownable() {
    let account = setup();
    // Deploy LINK token
    let calldata = link_deploy_args(contract_address_const::<123>(), // minter
     account // owner
    );

    let (linkAddr, _) = declare("LinkToken").unwrap().deploy(@calldata).unwrap();

    should_implement_ownable(linkAddr, account);
}

#[test]
#[should_panic(expected: ('minter is 0',))]
fn test_constructor_zero_address() {
    let sender = setup();
    let mut state = STATE();

    LinkToken::constructor(ref state, 0, 0, 0, 0, Zeroable::zero(), Zeroable::zero(), sender, 0);
}

#[test]
fn test_constructor() {
    let sender = setup();
    let mut state = STATE();
    LinkToken::constructor(ref state, 0, 0, 0, 0, Zeroable::zero(), sender, sender, 0);

    assert(Minter::minter(@state) == sender, 'minter valid');
    assert(state.erc20.name() == "ChainLink Token", 'name valid');
    assert(state.erc20.symbol() == "LINK", 'symbol valid');
}

#[test]
fn test_permissioned_mint_from_minter() {
    let sender = setup();
    let mut state = STATE();
    LinkToken::constructor(ref state, 0, 0, 0, 0, Zeroable::zero(), sender, sender, 0);
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
    LinkToken::constructor(ref state, 0, 0, 0, 0, Zeroable::zero(), minter, sender, 0);
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
    LinkToken::constructor(ref state, 0, 0, 0, 0, Zeroable::zero(), sender, sender, 0);
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
    LinkToken::constructor(ref state, 0, 0, 0, 0, Zeroable::zero(), minter, sender, 0);
    let to = contract_address_const::<908>();

    let amount: felt252 = 3000;
    MintableToken::permissioned_burn(ref state, to, amount.into());
}

#[test]
#[should_panic(expected: ('Caller is not the owner',))]
fn test_upgrade_non_owner() {
    let sender = setup();
    let mut state = STATE();
    LinkToken::constructor(
        ref state, 0, 0, 0, 0, Zeroable::zero(), sender, contract_address_const::<111>(), 0
    );

    UpgradeableImpl::upgrade(ref state, class_hash_const::<123>());
}

#[test]
#[should_panic(expected: ('Caller is not the owner',))]
fn test_set_minter_non_owner() {
    let sender = setup();
    let mut state = STATE();
    LinkToken::constructor(
        ref state, 0, 0, 0, 0, Zeroable::zero(), sender, contract_address_const::<111>(), 0
    );

    Minter::set_minter(ref state, contract_address_const::<123>())
}

#[test]
#[should_panic(expected: ('is minter already',))]
fn test_set_minter_already() {
    let sender = setup();
    let mut state = STATE();

    let minter = contract_address_const::<111>();
    LinkToken::constructor(ref state, 0, 0, 0, 0, Zeroable::zero(), minter, sender, 0);

    Minter::set_minter(ref state, minter);
}

#[test]
fn test_set_minter_success() {
    let sender = setup();
    let mut state = STATE();

    let minter = contract_address_const::<111>();
    LinkToken::constructor(ref state, 0, 0, 0, 0, Zeroable::zero(), minter, sender, 0);

    let new_minter = contract_address_const::<222>();
    Minter::set_minter(ref state, new_minter);

    assert(new_minter == Minter::minter(@state), 'new minter should be 222');
}

