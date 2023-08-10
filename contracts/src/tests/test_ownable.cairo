use starknet::contract_address_const;
use starknet::ContractAddress;
use starknet::testing::set_caller_address;
use starknet::testing::set_contract_address;
use zeroable::Zeroable;
use chainlink::libraries::ownable::{Ownable, IOwnable, IOwnableDispatcher, IOwnableDispatcherTrait};
use chainlink::libraries::ownable::Ownable::OwnableImpl;

fn STATE() -> Ownable::ContractState {
    Ownable::contract_state_for_testing()
}

fn setup() -> (ContractAddress, ContractAddress) {
    let owner: ContractAddress = contract_address_const::<1>();
    let other_user: ContractAddress = contract_address_const::<2>();
    set_caller_address(owner);
    (owner, other_user)
}

#[test]
#[available_gas(2000000)]
fn test_assert_only_owner() {
    let (owner, _) = setup();

    let mut state = STATE();
    Ownable::constructor(ref state, owner);
    Ownable::assert_only_owner(@state);
}

#[test]
#[available_gas(2000000)]
#[should_panic]
fn test_assert_only_owner_panics_if_not_owner() {
    let (owner, other_user) = setup();
    let mut state = STATE();

    Ownable::constructor(ref state, owner);
    set_caller_address(other_user);
    Ownable::assert_only_owner(@state);
}

#[test]
#[available_gas(2000000)]
fn test_owner() {
    let (owner, _) = setup();
    let mut state = STATE();

    Ownable::constructor(ref state, owner);

    assert(owner == OwnableImpl::owner(@state), 'should equal owner');
}

#[test]
#[available_gas(2000000)]
fn test_transfer_ownership() {
    let (owner, other_user) = setup();
    let mut state = STATE();

    Ownable::constructor(ref state, owner);
    OwnableImpl::transfer_ownership(ref state, other_user);

    assert(other_user == OwnableImpl::proposed_owner(@state), 'should equal proposed owner');
}

#[test]
#[available_gas(2000000)]
#[should_panic]
fn test_transfer_ownership_panics_if_zero_address() {
    let (owner, other_user) = setup();
    let mut state = STATE();

    Ownable::constructor(ref state, owner);
    OwnableImpl::transfer_ownership(ref state, Zeroable::zero());
}

#[test]
#[available_gas(2000000)]
#[should_panic]
fn test_transfer_ownership_panics_if_not_owner() {
    let (owner, other_user) = setup();
    let mut state = STATE();

    Ownable::constructor(ref state, owner);
    set_caller_address(other_user);
    OwnableImpl::transfer_ownership(ref state, other_user);
}

#[test]
#[available_gas(2000000)]
fn test_accept_ownership() {
    let (owner, other_user) = setup();
    let mut state = STATE();

    Ownable::constructor(ref state, owner);
    OwnableImpl::transfer_ownership(ref state, other_user);
    set_caller_address(other_user);
    OwnableImpl::accept_ownership(ref state);

    assert(OwnableImpl::owner(@state) == other_user, 'failed to accept ownership');
}

#[test]
#[available_gas(2000000)]
#[should_panic]
fn test_accept_ownership_panics_if_not_proposed_owner() {
    let (owner, other_user) = setup();
    let mut state = STATE();

    Ownable::constructor(ref state, owner);
    OwnableImpl::transfer_ownership(ref state, other_user);

    set_caller_address(contract_address_const::<3>());
    OwnableImpl::accept_ownership(ref state);
}

#[test]
#[available_gas(2000000)]
fn test_renounce_ownership() {
    let (owner, _) = setup();
    let mut state = STATE();

    set_caller_address(owner);
    Ownable::constructor(ref state, owner);
    OwnableImpl::renounce_ownership(ref state);

    assert(OwnableImpl::owner(@state).is_zero(), 'owner not 0 after renounce');
}

#[test]
#[available_gas(2000000)]
#[should_panic]
fn test_renounce_ownership_panics_if_not_owner() {
    let (owner, other_user) = setup();
    let mut state = STATE();

    Ownable::constructor(ref state, owner);
    set_caller_address(other_user);
    OwnableImpl::renounce_ownership(ref state);
}
//
// General ownable contract tests
//

fn should_implement_ownable(contract_addr: ContractAddress, owner: ContractAddress) {
    let contract = IOwnableDispatcher { contract_address: contract_addr };
    let acc2: ContractAddress = contract_address_const::<2222>();

    // check owner is set correctly
    assert(owner == contract.owner(), 'owner does not match');

    // transfer ownership - check owner unchanged and proposed owner set correctly
    set_contract_address(owner); // required to call contract as owner
    let caller = starknet::get_caller_address();
    contract.transfer_ownership(acc2);
    assert(owner == contract.owner(), 'owner should remain unchanged');
    assert(acc2 == contract.proposed_owner(), 'acc2 should be proposed owner');

    // accept ownership - check owner changed and proposed owner set to zero
    set_contract_address(acc2); // required to call function as acc2
    contract.accept_ownership();
    assert(contract.owner() == acc2, 'failed to change ownership');
    assert(contract.proposed_owner().is_zero(), 'proposed owner should be zero');

    // renounce ownership
    contract.renounce_ownership();
    assert(contract.owner().is_zero(), 'owner not 0 after renounce');
}

