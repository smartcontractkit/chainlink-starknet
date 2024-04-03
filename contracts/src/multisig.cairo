use array::ArrayTrait;
use option::OptionTrait;
use starknet::ContractAddress;
use starknet::class_hash::ClassHash;

fn assert_unique_values<
    T, impl TCopy: Copy<T>, impl TDrop: Drop<T>, impl TPartialEq: PartialEq<T>,
>(
    a: @Array::<T>
) {
    let len = a.len();
    _assert_unique_values_loop(a, len, 0_usize, 1_usize);
}

fn _assert_unique_values_loop<
    T, impl TCopy: Copy<T>, impl TDrop: Drop<T>, impl TPartialEq: PartialEq<T>,
>(
    a: @Array::<T>, len: usize, j: usize, k: usize
) {
    if j >= len {
        return ();
    }
    if k >= len {
        _assert_unique_values_loop(a, len, j + 1_usize, j + 2_usize);
        return ();
    }
    let j_val = *a.at(j);
    let k_val = *a.at(k);
    assert(j_val != k_val, 'duplicate values');
    _assert_unique_values_loop(a, len, j, k + 1_usize);
}

#[derive(Copy, Drop, Serde, starknet::Store)]
struct Transaction {
    to: ContractAddress,
    function_selector: felt252,
    calldata_len: usize,
    executed: bool,
    confirmations: usize,
}

#[starknet::interface]
trait IMultisig<TContractState> {
    fn is_signer(self: @TContractState, address: ContractAddress) -> bool;
    fn get_signers_len(self: @TContractState) -> usize;
    fn get_signers(self: @TContractState) -> Array<ContractAddress>;
    fn get_threshold(self: @TContractState) -> usize;
    fn get_transactions_len(self: @TContractState) -> u128;
    fn is_confirmed(self: @TContractState, nonce: u128, signer: ContractAddress) -> bool;
    fn is_executed(self: @TContractState, nonce: u128) -> bool;
    fn get_transaction(self: @TContractState, nonce: u128) -> (Transaction, Array::<felt252>);
    fn submit_transaction(
        ref self: TContractState,
        to: ContractAddress,
        function_selector: felt252,
        calldata: Array<felt252>
    );
    fn confirm_transaction(ref self: TContractState, nonce: u128);
    fn revoke_confirmation(ref self: TContractState, nonce: u128);
    fn execute_transaction(ref self: TContractState, nonce: u128) -> Array<felt252>;
    fn set_threshold(ref self: TContractState, threshold: usize);
    fn set_signers(ref self: TContractState, signers: Array<ContractAddress>);
    fn set_signers_and_threshold(
        ref self: TContractState, signers: Array<ContractAddress>, threshold: usize
    );
}

#[starknet::contract]
mod Multisig {
    use super::assert_unique_values;
    use super::{Transaction};

    use traits::Into;
    use traits::TryInto;
    use zeroable::Zeroable;
    use array::ArrayTrait;
    use array::ArrayTCloneImpl;
    use option::OptionTrait;

    use starknet::ContractAddress;
    use starknet::ContractAddressIntoFelt252;
    use starknet::Felt252TryIntoContractAddress;
    use starknet::StorageBaseAddress;
    use starknet::SyscallResult;
    use starknet::SyscallResultTrait;
    use starknet::call_contract_syscall;
    use starknet::get_caller_address;
    use starknet::get_contract_address;
    use starknet::storage_address_from_base_and_offset;
    use starknet::storage_read_syscall;
    use starknet::storage_write_syscall;
    use starknet::class_hash::ClassHash;

    use chainlink::libraries::type_and_version::ITypeAndVersion;
    use chainlink::libraries::upgradeable::{Upgradeable, IUpgradeable};

    #[event]
    #[derive(Drop, starknet::Event)]
    enum Event {
        TransactionSubmitted: TransactionSubmitted,
        TransactionConfirmed: TransactionConfirmed,
        ConfirmationRevoked: ConfirmationRevoked,
        TransactionExecuted: TransactionExecuted,
        SignersSet: SignersSet,
        ThresholdSet: ThresholdSet,
    }

    #[derive(Drop, starknet::Event)]
    struct TransactionSubmitted {
        #[key]
        signer: ContractAddress,
        #[key]
        nonce: u128,
        #[key]
        to: ContractAddress
    }

    #[derive(Drop, starknet::Event)]
    struct TransactionConfirmed {
        #[key]
        signer: ContractAddress,
        #[key]
        nonce: u128
    }

    #[derive(Drop, starknet::Event)]
    struct ConfirmationRevoked {
        #[key]
        signer: ContractAddress,
        #[key]
        nonce: u128
    }

    #[derive(Drop, starknet::Event)]
    struct TransactionExecuted {
        #[key]
        executor: ContractAddress,
        #[key]
        nonce: u128
    }

    #[derive(Drop, starknet::Event)]
    struct SignersSet {
        #[key]
        signers: Array<ContractAddress>
    }

    #[derive(Drop, starknet::Event)]
    struct ThresholdSet {
        #[key]
        threshold: usize
    }

    #[storage]
    struct Storage {
        _threshold: usize,
        _signers: LegacyMap<usize, ContractAddress>,
        _is_signer: LegacyMap<ContractAddress, bool>,
        _signers_len: usize,
        _tx_valid_since: u128,
        _next_nonce: u128,
        _transactions: LegacyMap<u128, Transaction>,
        _transaction_calldata: LegacyMap<(u128, usize), felt252>,
        _is_confirmed: LegacyMap<(u128, ContractAddress), bool>,
    }

    #[constructor]
    fn constructor(ref self: ContractState, signers: Array<ContractAddress>, threshold: usize) {
        let signers_len = signers.len();
        self._require_valid_threshold(threshold, signers_len);
        self._set_signers(signers, signers_len);
        self._set_threshold(threshold);
    }

    #[abi(embed_v0)]
    impl TypeAndVersionImpl of ITypeAndVersion<ContractState> {
        fn type_and_version(self: @ContractState,) -> felt252 {
            'Multisig 1.0.0'
        }
    }

    #[abi(embed_v0)]
    impl UpgradeableImpl of IUpgradeable<ContractState> {
        fn upgrade(ref self: ContractState, new_impl: ClassHash) {
            self._require_multisig();
            Upgradeable::upgrade(new_impl)
        }
    }

    #[abi(embed_v0)]
    impl MultisigImpl of super::IMultisig<ContractState> {
        /// Views

        fn is_signer(self: @ContractState, address: ContractAddress) -> bool {
            self._is_signer.read(address)
        }

        fn get_signers_len(self: @ContractState,) -> usize {
            self._signers_len.read()
        }

        fn get_signers(self: @ContractState) -> Array<ContractAddress> {
            let signers_len = self._signers_len.read();
            let mut signers = ArrayTrait::new();
            self._get_signers_range(0_usize, signers_len, ref signers);
            signers
        }

        fn get_threshold(self: @ContractState,) -> usize {
            self._threshold.read()
        }

        fn get_transactions_len(self: @ContractState,) -> u128 {
            self._next_nonce.read()
        }

        fn is_confirmed(self: @ContractState, nonce: u128, signer: ContractAddress) -> bool {
            self._is_confirmed.read((nonce, signer))
        }

        fn is_executed(self: @ContractState, nonce: u128) -> bool {
            let transaction = self._transactions.read(nonce);
            transaction.executed
        }

        fn get_transaction(self: @ContractState, nonce: u128) -> (Transaction, Array::<felt252>) {
            let transaction = self._transactions.read(nonce);

            let mut calldata = ArrayTrait::new();
            let calldata_len = transaction.calldata_len;
            self._get_transaction_calldata_range(nonce, 0_usize, calldata_len, ref calldata);

            (transaction, calldata)
        }

        /// Externals

        fn submit_transaction(
            ref self: ContractState,
            to: ContractAddress,
            function_selector: felt252,
            calldata: Array<felt252>,
        ) {
            self._require_signer();

            let nonce = self._next_nonce.read();
            let calldata_len = calldata.len();

            let transaction = Transaction {
                to: to,
                function_selector: function_selector,
                calldata_len: calldata_len,
                executed: false,
                confirmations: 0_usize
            };
            self._transactions.write(nonce, transaction);

            self._set_transaction_calldata_range(nonce, 0_usize, calldata_len, @calldata);

            let caller = get_caller_address();
            self
                .emit(
                    Event::TransactionSubmitted(
                        TransactionSubmitted { signer: caller, nonce: nonce, to: to }
                    )
                );
            self._next_nonce.write(nonce + 1_u128);
        }

        fn confirm_transaction(ref self: ContractState, nonce: u128) {
            self._require_signer();
            self._require_tx_exists(nonce);
            self._require_tx_valid(nonce);
            self._require_not_executed(nonce);
            self._require_not_confirmed(nonce);

            // TODO: write a single field instead of the whole transaction?
            let mut transaction = self._transactions.read(nonce);
            transaction.confirmations += 1_usize;
            self._transactions.write(nonce, transaction);

            let caller = get_caller_address();
            self._is_confirmed.write((nonce, caller), true);

            self
                .emit(
                    Event::TransactionConfirmed(
                        TransactionConfirmed { signer: caller, nonce: nonce }
                    )
                );
        }

        fn revoke_confirmation(ref self: ContractState, nonce: u128) {
            self._require_signer();
            self._require_tx_exists(nonce);
            self._require_tx_valid(nonce);
            self._require_not_executed(nonce);
            self._require_confirmed(nonce);

            // TODO: write a single field instead of the whole transaction?
            let mut transaction = self._transactions.read(nonce);
            transaction.confirmations -= 1_usize;
            self._transactions.write(nonce, transaction);

            let caller = get_caller_address();
            self._is_confirmed.write((nonce, caller), false);

            self
                .emit(
                    Event::ConfirmationRevoked(ConfirmationRevoked { signer: caller, nonce: nonce })
                );
        }

        fn execute_transaction(ref self: ContractState, nonce: u128) -> Array<felt252> {
            self._require_signer();
            self._require_tx_exists(nonce);
            self._require_tx_valid(nonce);
            self._require_not_executed(nonce);

            let mut transaction = self._transactions.read(nonce);

            let threshold = self._threshold.read();
            assert(threshold <= transaction.confirmations, 'more confirmations required');

            let mut calldata = ArrayTrait::new();
            let calldata_len = transaction.calldata_len;
            self._get_transaction_calldata_range(nonce, 0_usize, calldata_len, ref calldata);

            transaction.executed = true;
            self._transactions.write(nonce, transaction);

            let caller = get_caller_address();
            self
                .emit(
                    Event::TransactionExecuted(
                        TransactionExecuted { executor: caller, nonce: nonce }
                    )
                );

            let response = call_contract_syscall(
                transaction.to, transaction.function_selector, calldata.span()
            )
                .unwrap_syscall();

            // TODO: this shouldn't be necessary. call_contract_syscall returns a Span<felt252>, which
            // is a serialized result, but returning a Span<felt252> results in an error:
            //
            // Trait has no implementation in context: core::serde::Serde::<core::array::Span::<core::felt252>>
            //
            // Cairo docs also have an example that returns a Span<felt252>:
            // https://github.com/starkware-libs/cairo/blob/fe425d0893ff93a936bb3e8bbbac771033074bdb/docs/reference/src/components/cairo/modules/language_constructs/pages/contracts.adoc#L226
            ArrayTCloneImpl::clone(response.snapshot)
        }

        fn set_threshold(ref self: ContractState, threshold: usize) {
            self._require_multisig();

            let signers_len = self._signers_len.read();
            self._require_valid_threshold(threshold, signers_len);

            self._update_tx_valid_since();

            self._set_threshold(threshold);
        }

        fn set_signers(ref self: ContractState, signers: Array<ContractAddress>) {
            self._require_multisig();

            self._update_tx_valid_since();

            let signers_len = signers.len();
            self._set_signers(signers, signers_len);

            let threshold = self._threshold.read();

            if signers_len < threshold {
                self._require_valid_threshold(signers_len, signers_len);
                self._set_threshold(signers_len);
            }
        }

        fn set_signers_and_threshold(
            ref self: ContractState, signers: Array<ContractAddress>, threshold: usize
        ) {
            self._require_multisig();

            let signers_len = signers.len();
            self._require_valid_threshold(threshold, signers_len);

            self._update_tx_valid_since();

            self._set_signers(signers, signers_len);
            self._set_threshold(threshold);
        }
    }

    /// Internals
    #[generate_trait]
    impl InternalImpl of InternalTrait {
        fn _set_signers(
            ref self: ContractState, signers: Array<ContractAddress>, signers_len: usize
        ) {
            self._require_unique_signers(@signers);

            let old_signers_len = self._signers_len.read();
            self._clean_signers_range(0_usize, old_signers_len);

            self._signers_len.write(signers_len);
            self._set_signers_range(0_usize, signers_len, @signers);

            self.emit(Event::SignersSet(SignersSet { signers: signers }));
        }

        fn _clean_signers_range(ref self: ContractState, index: usize, len: usize) {
            if index >= len {
                return ();
            }

            let signer = self._signers.read(index);
            self._is_signer.write(signer, false);
            self._signers.write(index, Zeroable::zero());

            self._clean_signers_range(index + 1_usize, len);
        }

        fn _set_signers_range(
            ref self: ContractState, index: usize, len: usize, signers: @Array<ContractAddress>
        ) {
            if index >= len {
                return ();
            }

            let signer = *signers.at(index);
            self._signers.write(index, signer);
            self._is_signer.write(signer, true);

            self._set_signers_range(index + 1_usize, len, signers);
        }

        fn _get_signers_range(
            self: @ContractState, index: usize, len: usize, ref signers: Array<ContractAddress>
        ) {
            if index >= len {
                return ();
            }

            let signer = self._signers.read(index);
            signers.append(signer);

            self._get_signers_range(index + 1_usize, len, ref signers);
        }

        fn _set_transaction_calldata_range(
            ref self: ContractState,
            nonce: u128,
            index: usize,
            len: usize,
            calldata: @Array<felt252>
        ) {
            if index >= len {
                return ();
            }

            let calldata_arg = *calldata.at(index);
            self._transaction_calldata.write((nonce, index), calldata_arg);

            self._set_transaction_calldata_range(nonce, index + 1_usize, len, calldata);
        }

        fn _get_transaction_calldata_range(
            self: @ContractState,
            nonce: u128,
            index: usize,
            len: usize,
            ref calldata: Array<felt252>
        ) {
            if index >= len {
                return ();
            }

            let calldata_arg = self._transaction_calldata.read((nonce, index));
            calldata.append(calldata_arg);

            self._get_transaction_calldata_range(nonce, index + 1_usize, len, ref calldata);
        }

        fn _set_threshold(ref self: ContractState, threshold: usize) {
            self._threshold.write(threshold);
            self.emit(Event::ThresholdSet(ThresholdSet { threshold: threshold }));
        }

        fn _update_tx_valid_since(ref self: ContractState) {
            let tx_valid_since = self._next_nonce.read();
            self._tx_valid_since.write(tx_valid_since);
        }

        fn _require_signer(self: @ContractState) {
            let caller = get_caller_address();
            let is_signer = self._is_signer.read(caller);
            assert(is_signer, 'invalid signer');
        }

        fn _require_tx_exists(self: @ContractState, nonce: u128) {
            let next_nonce = self._next_nonce.read();
            assert(nonce < next_nonce, 'transaction does not exist');
        }

        fn _require_not_executed(self: @ContractState, nonce: u128) {
            let transaction = self._transactions.read(nonce);
            assert(!transaction.executed, 'transaction already executed');
        }

        fn _require_not_confirmed(self: @ContractState, nonce: u128) {
            let caller = get_caller_address();
            let is_confirmed = self._is_confirmed.read((nonce, caller));
            assert(!is_confirmed, 'transaction already confirmed');
        }

        fn _require_confirmed(self: @ContractState, nonce: u128) {
            let caller = get_caller_address();
            let is_confirmed = self._is_confirmed.read((nonce, caller));
            assert(is_confirmed, 'transaction not confirmed');
        }

        fn _require_unique_signers(self: @ContractState, signers: @Array<ContractAddress>) {
            assert_unique_values(signers);
        }

        fn _require_tx_valid(self: @ContractState, nonce: u128) {
            let tx_valid_since = self._tx_valid_since.read();
            assert(tx_valid_since <= nonce, 'transaction invalid');
        }

        fn _require_multisig(self: @ContractState) {
            let caller = get_caller_address();
            let contract = get_contract_address();
            assert(caller == contract, 'only multisig allowed');
        }

        fn _require_valid_threshold(self: @ContractState, threshold: usize, signers_len: usize) {
            if threshold == 0_usize {
                if signers_len == 0_usize {
                    return ();
                }
            }

            assert(threshold >= 1_usize, 'invalid threshold, too small');
            assert(threshold <= signers_len, 'invalid threshold, too large');
        }
    }
}
