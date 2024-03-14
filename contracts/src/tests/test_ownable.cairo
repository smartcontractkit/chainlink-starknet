use starknet::contract_address_const;
use starknet::ContractAddress;
use starknet::testing::set_caller_address;
use starknet::testing::set_contract_address;
use zeroable::Zeroable;

use openzeppelin::access::ownable::interface::{
    IOwnableTwoStep, IOwnableTwoStepDispatcher, IOwnableTwoStepDispatcherTrait
};

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

