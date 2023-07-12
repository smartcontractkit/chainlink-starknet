// copied from https://github.com/OpenZeppelin/cairo-contracts/pull/616

use serde::Serde;
use array::ArrayTrait;
use starknet::ContractAddress;
use starknet::SyscallResultTrait;

const ERC165_ACCOUNT_ID: u32 = 0xa66bd575_u32;
const ERC1271_VALIDATED: u32 = 0x1626ba7e_u32;

const TRANSACTION_VERSION: felt252 = 1;
const QUERY_VERSION: felt252 =
    340282366920938463463374607431768211457; // 2**128 + TRANSACTION_VERSION

#[inline(always)]
fn check_gas() {
    match gas::withdraw_gas() {
        Option::Some(_) => {},
        Option::None(_) => {
            let mut data = ArrayTrait::new();
            data.append('Out of gas');
            panic(data);
        },
    }
}

#[derive(Serde, Drop)]
struct Call {
    to: ContractAddress,
    selector: felt252,
    calldata: Array<felt252>
}

#[starknet::interface]
trait IAccount {
    fn __validate__(calls: Array<Call>) -> felt252;
    fn __validate_declare__(class_hash: felt252) -> felt252;
}

#[account_contract]
mod Account {
    use array::SpanTrait;
    use array::ArrayTrait;
    use box::BoxTrait;
    use option::OptionTrait;
    use zeroable::Zeroable;
    use ecdsa::check_ecdsa_signature;
    use starknet::get_tx_info;
    use starknet::get_caller_address;
    use starknet::get_contract_address;

    use super::Call;
    use super::ERC165_ACCOUNT_ID;
    use super::ERC1271_VALIDATED;
    use super::TRANSACTION_VERSION;
    use super::QUERY_VERSION;
    use super::check_gas;

    use chainlink::account::erc165::ERC165;

    //
    // Storage and Constructor
    //

    #[storage]
    struct Storage {
        public_key: felt252, 
    }

    #[constructor]
    fn constructor(ref self: ContractState, _public_key: felt252) {
        ERC165::register_interface(ERC165_ACCOUNT_ID);
        self.public_key.write(_public_key);
    }

    //
    // Externals
    //

    // todo: fix Span serde
    // #[external(v0)]
    fn __execute__(mut calls: Array<Call>) -> Array<Span<felt252>> {
        // avoid calls from other contracts
        // https://github.com/OpenZeppelin/cairo-contracts/issues/344
        let sender = get_caller_address();
        assert(sender.is_zero(), 'Account: invalid caller');

        // check tx version
        let tx_info = get_tx_info().unbox();
        let version = tx_info.version;
        if version != TRANSACTION_VERSION { // > operator not defined for felt252
            assert(version == QUERY_VERSION, 'Account: invalid tx version');
        }

        _execute_calls(calls, ArrayTrait::new())
    }

    #[external(v0)]
    fn __validate__(mut calls: Array<Call>) -> felt252 {
        _validate_transaction()
    }

    #[external(v0)]
    fn __validate_declare__(class_hash: felt252) -> felt252 {
        _validate_transaction()
    }

    #[external(v0)]
    fn __validate_deploy__(
        class_hash: felt252, contract_address_salt: felt252, _public_key: felt252
    ) -> felt252 {
        _validate_transaction()
    }

    #[external(v0)]
    fn set_public_key(new_public_key: felt252) {
        _assert_only_self();
        public_key::write(new_public_key);
    }

    //
    // View
    //

    #[view]
    fn get_public_key() -> felt252 {
        public_key::read()
    }

    // todo: fix Span serde
    // #[view]
    fn is_valid_signature(message: felt252, signature: Span<felt252>) -> u32 {
        if _is_valid_signature(message, signature) {
            ERC1271_VALIDATED
        } else {
            0_u32
        }
    }

    #[view]
    fn supports_interface(interface_id: u32) -> bool {
        ERC165::supports_interface(interface_id)
    }

    //
    // Internals
    //

    #[internal]
    fn _assert_only_self() {
        let caller = get_caller_address();
        let self = get_contract_address();
        assert(self == caller, 'Account: unauthorized');
    }

    #[internal]
    fn _validate_transaction() -> felt252 {
        let tx_info = get_tx_info().unbox();
        let tx_hash = tx_info.transaction_hash;
        let signature = tx_info.signature;
        assert(_is_valid_signature(tx_hash, signature), 'Account: invalid signature');
        starknet::VALIDATED
    }

    #[internal]
    fn _is_valid_signature(message: felt252, signature: Span<felt252>) -> bool {
        let valid_length = signature.len() == 2_u32;

        valid_length
            & check_ecdsa_signature(
                message, public_key::read(), *signature.at(0_u32), *signature.at(1_u32)
            )
    }

    #[internal]
    fn _execute_calls(
        mut calls: Array<Call>, mut res: Array<Span<felt252>>
    ) -> Array<Span<felt252>> {
        check_gas();
        match calls.pop_front() {
            Option::Some(call) => {
                let _res = _execute_single_call(call);
                res.append(_res);
                return _execute_calls(calls, res);
            },
            Option::None(_) => {
                return res;
            },
        }
    }

    #[internal]
    fn _execute_single_call(mut call: Call) -> Span<felt252> {
        let Call{to, selector, calldata } = call;
        starknet::call_contract_syscall(to, selector, calldata.span()).unwrap_syscall()
    }
}
