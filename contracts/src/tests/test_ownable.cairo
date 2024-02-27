use starknet::contract_address_const;
use starknet::ContractAddress;
use starknet::testing::set_caller_address;
use starknet::testing::set_contract_address;
use zeroable::Zeroable;

use openzeppelin::access::ownable::ownable::OwnableComponent;
use openzeppelin::access::ownable::interface::{
    IOwnableTwoStep, IOwnableTwoStepDispatcher, IOwnableTwoStepDispatcherTrait
};
use OwnableComponent::InternalTrait;

#[starknet::contract]
mod OwnableMock {
    use super::OwnableComponent;
    use starknet::ContractAddress;

    component!(path: OwnableComponent, storage: ownable, event: OwnableEvent);

    #[abi(embed_v0)]
    impl OwnableImpl = OwnableComponent::OwnableTwoStepImpl<ContractState>;
    impl InternalImpl = OwnableComponent::InternalImpl<ContractState>;

    #[storage]
    struct Storage {
        #[substorage(v0)]
        ownable: OwnableComponent::Storage
    }

    #[event]
    #[derive(Drop, starknet::Event)]
    enum Event {
        #[flat]
        OwnableEvent: OwnableComponent::Event
    }

    #[constructor]
    fn constructor(ref self: ContractState, owner: ContractAddress) {
        self.ownable.initializer(owner);
    }
}

type ComponentState = OwnableComponent::ComponentState<OwnableMock::ContractState>;

fn STATE() -> ComponentState {
    OwnableComponent::component_state_for_testing()
}

fn setup() -> (ContractAddress, ContractAddress) {
    let owner: ContractAddress = contract_address_const::<1>();
    let other_user: ContractAddress = contract_address_const::<2>();
    set_caller_address(owner);
    (owner, other_user)
}

#[test]
fn test_assert_only_owner() {
    let (owner, _) = setup();

    let mut state = STATE();
    state.initializer(owner);
    state.assert_only_owner();
}

#[test]
#[should_panic]
fn test_assert_only_owner_panics_if_not_owner() {
    let (owner, other_user) = setup();
    let mut state = STATE();

    state.initializer(owner);
    set_caller_address(other_user);
    state.assert_only_owner();
}

#[test]
fn test_owner() {
    let (owner, _) = setup();
    let mut state = STATE();

    state.initializer(owner);

    assert(owner == state.owner(), 'should equal owner');
}

#[test]
fn test_transfer_ownership() {
    let (owner, other_user) = setup();
    let mut state = STATE();

    state.initializer(owner);
    state.transfer_ownership(other_user);

    assert(other_user == state.pending_owner(), 'should equal proposed owner');
}

#[test]
#[should_panic]
fn test_transfer_ownership_panics_if_not_owner() {
    let (owner, other_user) = setup();
    let mut state = STATE();

    state.initializer(owner);
    set_caller_address(other_user);
    state.transfer_ownership(other_user);
}

#[test]
fn test_accept_ownership() {
    let (owner, other_user) = setup();
    let mut state = STATE();

    state.initializer(owner);
    state.transfer_ownership(other_user);
    set_caller_address(other_user);
    state.accept_ownership();

    assert(state.owner() == other_user, 'failed to accept ownership');
}

#[test]
#[should_panic]
fn test_accept_ownership_panics_if_not_pending_owner() {
    let (owner, other_user) = setup();
    let mut state = STATE();

    state.initializer(owner);
    state.transfer_ownership(other_user);

    set_caller_address(contract_address_const::<3>());
    state.accept_ownership();
}

#[test]
fn test_renounce_ownership() {
    let (owner, _) = setup();
    let mut state = STATE();

    set_caller_address(owner);
    state.initializer(owner);
    state.renounce_ownership();

    assert(state.owner().is_zero(), 'owner not 0 after renounce');
}

#[test]
#[should_panic]
fn test_renounce_ownership_panics_if_not_owner() {
    let (owner, other_user) = setup();
    let mut state = STATE();

    state.initializer(owner);
    set_caller_address(other_user);
    state.renounce_ownership();
}
//
// General ownable contract tests
//

fn should_implement_ownable(contract_addr: ContractAddress, owner: ContractAddress) {
    let contract = IOwnableTwoStepDispatcher { contract_address: contract_addr };
    let acc2: ContractAddress = contract_address_const::<2222>();

    // check owner is set correctly
    assert(owner == contract.owner(), 'owner does not match');

    // transfer ownership - check owner unchanged and proposed owner set correctly
    set_contract_address(owner); // required to call contract as owner
    contract.transfer_ownership(acc2);
    assert(owner == contract.owner(), 'owner should remain unchanged');
    assert(acc2 == contract.pending_owner(), 'acc2 should be proposed owner');

    // accept ownership - check owner changed and proposed owner set to zero
    set_contract_address(acc2); // required to call function as acc2
    contract.accept_ownership();
    assert(contract.owner() == acc2, 'failed to change ownership');
    assert(contract.pending_owner().is_zero(), 'proposed owner should be zero');

    // renounce ownership
    contract.renounce_ownership();
    assert(contract.owner().is_zero(), 'owner not 0 after renounce');
}

