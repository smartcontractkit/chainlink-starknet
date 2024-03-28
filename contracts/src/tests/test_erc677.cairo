use starknet::ContractAddress;
use starknet::contract_address_const;
use starknet::testing::set_caller_address;
use starknet::syscalls::deploy_syscall;
use starknet::class_hash::Felt252TryIntoClassHash;

use array::ArrayTrait;
use traits::Into;
use traits::TryInto;
use zeroable::Zeroable;
use option::OptionTrait;
use core::result::ResultTrait;

use chainlink::token::mock::valid_erc667_receiver::ValidReceiver;
use chainlink::token::mock::invalid_erc667_receiver::InvalidReceiver;
use chainlink::libraries::token::erc677::ERC677Component;
use chainlink::libraries::token::erc677::ERC677Component::ERC677Impl;

#[starknet::interface]
trait MockInvalidReceiver<TContractState> {
    fn set_supports(ref self: TContractState, value: bool);
}

use chainlink::token::mock::valid_erc667_receiver::{
    MockValidReceiver, MockValidReceiverDispatcher, MockValidReceiverDispatcherTrait
};

// Ignored tests are dependent on upgrading our version of cairo to include this PR https://github.com/starkware-libs/cairo/pull/2912/files

fn setup() -> ContractAddress {
    let account: ContractAddress = contract_address_const::<1>();
    // Set account as default caller
    set_caller_address(account);
    account
}

fn setup_valid_receiver() -> (ContractAddress, MockValidReceiverDispatcher) {
    let calldata = ArrayTrait::new();
    let (address, _) = deploy_syscall(
        ValidReceiver::TEST_CLASS_HASH.try_into().unwrap(), 0, calldata.span(), false
    )
        .unwrap();
    let contract = MockValidReceiverDispatcher { contract_address: address };
    (address, contract)
}


fn setup_invalid_receiver() -> (ContractAddress, MockInvalidReceiverDispatcher) {
    let calldata = ArrayTrait::new();
    let (address, _) = deploy_syscall(
        InvalidReceiver::TEST_CLASS_HASH.try_into().unwrap(), 0, calldata.span(), false
    )
        .unwrap();
    let contract = MockInvalidReceiverDispatcher { contract_address: address };
    (address, contract)
}

type ComponentState =
    ERC677Component::ComponentState<chainlink::token::link_token::LinkToken::ContractState>;

fn transfer_and_call(receiver: ContractAddress) {
    let data = ArrayTrait::<felt252>::new();
    // have to send 0 because ERC20 is not initialized with starting supply when using this library by itself
    let mut state: ComponentState = ERC677Component::component_state_for_testing();
    state.transfer_and_call(receiver, u256 { high: 0, low: 0 }, data);
}

#[test]
#[should_panic(expected: ('ERC20: transfer to 0',))]
fn test_to_zero_address() {
    setup();
    transfer_and_call(Zeroable::zero());
}

#[test]
fn test_valid_transfer_and_call() {
    let sender = setup();
    let (receiver_address, receiver) = setup_valid_receiver();

    transfer_and_call(receiver_address);

    assert(receiver.verify() == sender, 'on_token_transfer called');
}

#[test]
#[should_panic(expected: ('ENTRYPOINT_NOT_FOUND',))]
fn test_invalid_receiver_supports_interface_true() {
    setup();
    let (receiver_address, receiver) = setup_invalid_receiver();

    receiver.set_supports(true);

    transfer_and_call(receiver_address);
}

#[test]
fn test_invalid_receiver_supports_interface_false() {
    setup();
    let (receiver_address, _) = setup_invalid_receiver();

    transfer_and_call(receiver_address);
}


#[test]
#[should_panic(expected: ('CONTRACT_NOT_DEPLOYED',))]
fn test_nonexistent_receiver() {
    setup();

    transfer_and_call(contract_address_const::<777>());
}

