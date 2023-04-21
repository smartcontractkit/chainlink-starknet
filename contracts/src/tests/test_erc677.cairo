use chainlink::libraries::token::erc677::ERC677;
use starknet::ContractAddress;
use starknet::contract_address_const;
use starknet::testing::set_caller_address;
use array::ArrayTrait;
use traits::Into;
use zeroable::Zeroable;


// todos are dependent on cairo-test support for contract calls
//  https://github.com/starkware-libs/cairo/blob/main/crates/cairo-lang-runner/src/casm_run/mod.rs#L466

fn setup() -> ContractAddress {
    let account: ContractAddress = contract_address_const::<1>();
    // Set account as default caller
    set_caller_address(account);
    account
}

// todo
fn setup_valid_receiver() -> ContractAddress {
    contract_address_const::<234>()
}

// todo
fn setup_invalid_receiver() -> ContractAddress {
    contract_address_const::<123>()
}

fn transfer_and_call(receiver: ContractAddress) {
    let data = ArrayTrait::<felt252>::new();
    // have to send 0 because ERC20 is not initialized when using this library by itself
    ERC677::transfer_and_call(receiver, u256 { high: 0, low: 0 }, data);
}

#[test]
#[available_gas(2000000)]
#[should_panic(expected = ('ERC20: transfer to 0', ))]
fn test_to_zero_address() {
    setup();
    transfer_and_call(Zeroable::zero());
}

// todo
fn test_valid_transfer_and_call() {
    setup();
    let receiver = setup_valid_receiver();

    transfer_and_call(receiver);
// assert Transfer event emited (not supported yet)
// assert that receiver's onTokenTransfer called was called
}

// todo
fn test_invalid_receiver() {
    setup();
    let receiver = setup_invalid_receiver();

    transfer_and_call(receiver);
// should panic
}

// todo
fn test_nonexistent_receiver() {
    setup();

    transfer_and_call(contract_address_const::<777>());
// should panic
}

