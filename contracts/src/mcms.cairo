use starknet::ContractAddress;
use starknet::{
    eth_signature::public_key_point_to_eth_address, EthAddress,
    secp256_trait::{
        Secp256Trait, Secp256PointTrait, recover_public_key, is_signature_entry_valid, Signature
    },
    secp256k1::Secp256k1Point, SyscallResult, SyscallResultTrait
};
use alexandria_bytes::{Bytes, BytesTrait};
use alexandria_encoding::sol_abi::sol_bytes::SolBytesTrait;
use alexandria_encoding::sol_abi::encode::SolAbiEncodeTrait;

#[starknet::interface]
trait IManyChainMultiSig<TContractState> {
    // todo: check length of byte array is 32
    fn set_root(
        ref self: TContractState,
        root: u256,
        valid_until: u32,
        metadata: RootMetadata,
        metadata_proof: Span<u256>,
        // note: v is a boolean and not uint8
        signatures: Array<Signature>
    );
    fn execute(ref self: TContractState, op: Op, proof: Span<u256>);
    // // todo: check length of group_quorums and group_parents
    fn set_config(
        ref self: TContractState,
        signer_addresses: Span<EthAddress>,
        signer_groups: Span<u8>,
        group_quorums: Span<u8>,
        group_parents: Span<u8>,
        clear_root: bool
    );
// fn get_config(self: @TContractState) -> Config;
// fn get_op_count(self: @TContractState) -> u64;
// fn get_root(self: @TContractState) -> (u256, u32);
// fn get_root_metadata(self: @TContractState) -> RootMetadata;
}

#[derive(Copy, Drop, Serde, starknet::Store)]
struct Signer {
    address: EthAddress,
    index: u8,
    group: u8
}

#[derive(Copy, Drop, Serde, starknet::Store)]
struct RootMetadata {
    chain_id: u256,
    multisig: ContractAddress,
    pre_op_count: u64,
    post_op_count: u64,
    override_previous_root: bool
}

// todo: maybe use copy
// todo: figure out how this works off-chain with MCMS since we have a new selector field here
#[derive(Copy, Drop, Serde)]
struct Op {
    chain_id: u256,
    multisig: ContractAddress,
    nonce: u64,
    to: ContractAddress,
    selector: felt252,
    data: Span<felt252>
}

// does not implement Storage trait because structs cannot support arrays or maps
#[derive(Copy, Drop, Serde)]
struct Config {
    signers: Span<Signer>,
    group_quorums: Span<u8>,
    group_parents: Span<u8>
}

#[derive(Copy, Drop, Serde, starknet::Store)]
struct ExpiringRootAndOpCount {
    root: u256,
    valid_until: u32,
    op_count: u64
}

// based of https://github.com/starkware-libs/cairo/blob/1b747da1ec7e43a6fd0c0a4cbce302616408bc72/corelib/src/starknet/eth_signature.cairo#L25
pub fn recover_eth_ecdsa(msg_hash: u256, signature: Signature) -> Result<EthAddress, felt252> {
    if !is_signature_entry_valid::<Secp256k1Point>(signature.r) {
        return Result::Err('Signature out of range');
    }
    if !is_signature_entry_valid::<Secp256k1Point>(signature.s) {
        return Result::Err('Signature out of range');
    }

    let public_key_point = recover_public_key::<Secp256k1Point>(:msg_hash, :signature).unwrap();
    // calculated eth address
    return Result::Ok(public_key_point_to_eth_address(:public_key_point));
}

pub fn to_u256(address: EthAddress) -> u256 {
    let temp: felt252 = address.into();
    temp.into()
}

pub fn verify_merkle_proof(proof: Span<u256>, root: u256, leaf: u256) -> bool {
    let mut computed_hash = leaf;

    let mut i = 0;

    while i < proof.len() {
        computed_hash = hash_pair(computed_hash, *proof.at(i));
        i += 1;
    };

    computed_hash == root
}

fn hash_pair(a: u256, b: u256) -> u256 {
    let (lower, higher) = if a < b {
        (a, b)
    } else {
        (b, a)
    };
    BytesTrait::new_empty().encode(lower).encode(higher).keccak()
}

#[starknet::contract]
mod ManyChainMultiSig {
    use core::starknet::SyscallResultTrait;
    use core::array::SpanTrait;
    use core::dict::Felt252Dict;
    use core::traits::PanicDestruct;
    use super::{
        ExpiringRootAndOpCount, Config, Signer, RootMetadata, Op, Signature, recover_eth_ecdsa,
        to_u256, verify_merkle_proof
    };
    use starknet::{
        EthAddress, EthAddressZeroable, EthAddressIntoFelt252, ContractAddress,
        call_contract_syscall
    };
    use starknet::eth_signature::is_eth_signature_valid;
    use alexandria_bytes::{Bytes, BytesTrait};
    use alexandria_encoding::sol_abi::sol_bytes::SolBytesTrait;
    use alexandria_encoding::sol_abi::encode::SolAbiEncodeTrait;

    const NUM_GROUPS: u8 = 32;
    const MAX_NUM_SIGNERS: u8 = 200;
    // keccak256("MANY_CHAIN_MULTI_SIG_DOMAIN_SEPARATOR_OP")
    const MANY_CHAIN_MULTI_SIG_DOMAIN_SEPARATOR_OP: u256 =
        0x08d275622006c4ca82d03f498e90163cafd53c663a48470c3b52ac8bfbd9f52c;
    // keccak256("MANY_CHAIN_MULTI_SIG_DOMAIN_SEPARATOR_METADATA")
    const MANY_CHAIN_MULTI_SIG_DOMAIN_SEPARATOR_METADATA: u256 =
        0xe6b82be989101b4eb519770114b997b97b3c8707515286748a871717f0e4ea1c;

    #[storage]
    struct Storage {
        s_signers: LegacyMap<EthAddress, Signer>,
        // begin s_config (defined in storage bc Config struct cannot support maps) 
        _s_config_signers_len: u8,
        _s_config_signers: LegacyMap<u8, Signer>,
        // no _s_config_group_len because there are always 32 groups
        _s_config_group_quorums: LegacyMap<u8, u8>,
        _s_config_group_parents: LegacyMap<u8, u8>,
        // end s_config
        s_seen_signed_hashes: LegacyMap<u256, bool>,
        s_expiring_root_and_op_count: ExpiringRootAndOpCount,
        s_root_metadata: RootMetadata
    }

    #[derive(Drop, starknet::Event)]
    struct NewRoot {
        #[key]
        root: u256,
        valid_until: u32,
        metadata: RootMetadata,
    }

    #[derive(Drop, starknet::Event)]
    struct OpExecuted {
        #[key]
        nonce: u64,
        to: ContractAddress,
        selector: felt252,
        data: Span<felt252>
    // no value because value is sent through ERC20 tokens, even the native STRK token
    }

    #[event]
    #[derive(Drop, starknet::Event)]
    enum Event {
        NewRoot: NewRoot,
        OpExecuted: OpExecuted,
    }


    #[abi(embed_v0)]
    impl ManyChainMultiSigImpl of super::IManyChainMultiSig<ContractState> {
        fn set_root(
            ref self: ContractState,
            root: u256,
            valid_until: u32,
            metadata: RootMetadata,
            metadata_proof: Span<u256>,
            // note: v is a boolean and not uint8
            mut signatures: Array<Signature>
        ) {
            let encoded_root: Bytes = BytesTrait::new_empty().encode(root).encode(valid_until);

            let mut eip_191_msg: Bytes = BytesTrait::new_empty();
            eip_191_msg.append_felt252('\x19Ethereum Signed Message:\n32');
            eip_191_msg.append_u256(encoded_root.keccak());
            let msg_hash = eip_191_msg.keccak();

            assert(!self.s_seen_signed_hashes.read(msg_hash), 'signed hash already seen');

            let prev_address = EthAddressZeroable::zero();
            let mut group_vote_counts: Felt252Dict<u8> = Default::default();
            while let Option::Some(signature) = signatures
                .pop_front() {
                    let signer_address = match recover_eth_ecdsa(msg_hash, signature) {
                        Result::Ok(signer_address) => signer_address,
                        Result::Err(e) => panic_with_felt252(e),
                    };

                    assert(
                        to_u256(prev_address) < to_u256(signer_address),
                        'signer address must increase'
                    );

                    let signer = self.s_signers.read(signer_address);
                    assert(signer.address == signer_address, 'invalid signer');

                    let mut group = signer.group;
                    loop {
                        // todo: may be unnecessary assert
                        assert(group < NUM_GROUPS, 'invalid group number');
                        let counts = group_vote_counts.get(group.into());
                        group_vote_counts.insert(group.into(), counts + 1);
                        if counts + 1 != self._s_config_group_quorums.read(group) {
                            break;
                        }
                        if group == 0 {
                            // reached root
                            break;
                        }
                        group = self._s_config_group_parents.read(group)
                    };
                };

            let root_group_quorum = self._s_config_group_quorums.read(0);
            assert(root_group_quorum > 0, 'missing config');
            assert(group_vote_counts.get(0) >= root_group_quorum, 'insufficient signers');
            assert(
                valid_until.into() >= starknet::info::get_block_timestamp(),
                'valid until has passed'
            );
            // verify metadataProof
            // todo: make sure this is the right way to encode the struct
            let encoded_metadata: Bytes = BytesTrait::new_empty()
                .encode(MANY_CHAIN_MULTI_SIG_DOMAIN_SEPARATOR_METADATA)
                .encode(valid_until)
                .encode(metadata.chain_id)
                .encode(metadata.multisig)
                .encode(metadata.pre_op_count)
                .encode(metadata.post_op_count)
                .encode(metadata.override_previous_root);

            let hashed_leaf = encoded_metadata.keccak();
            assert(
                verify_merkle_proof(metadata_proof, root, hashed_leaf), 'proof verification failed'
            );

            // maybe move to beginning of function
            assert(
                starknet::info::get_tx_info().unbox().chain_id.into() == metadata.chain_id,
                'wrong chain id'
            );
            assert(
                starknet::info::get_contract_address() == metadata.multisig,
                'wrong multisig address'
            );

            let op_count = self.s_expiring_root_and_op_count.read().op_count;
            let current_root_metadata = self.s_root_metadata.read();
            assert(
                op_count == current_root_metadata.post_op_count
                    || current_root_metadata.override_previous_root,
                'expect pending operations'
            );
            assert(op_count == metadata.pre_op_count, 'wrong pre-operation count');
            assert(metadata.pre_op_count <= metadata.post_op_count, 'wrong post-operation count');

            self.s_seen_signed_hashes.write(msg_hash, true);
            self
                .s_expiring_root_and_op_count
                .write(
                    ExpiringRootAndOpCount {
                        root: root, valid_until: valid_until, op_count: metadata.pre_op_count
                    }
                );

            self
                .emit(
                    Event::NewRoot(
                        NewRoot { root: root, valid_until: valid_until, metadata: metadata, }
                    )
                );
        }

        fn execute(ref self: ContractState, op: Op, proof: Span<u256>) {
            let current_expiring_root_and_op_count = self.s_expiring_root_and_op_count.read();

            assert(
                self
                    .s_root_metadata
                    .read()
                    .post_op_count > current_expiring_root_and_op_count
                    .op_count,
                'post-operation count reached'
            );

            assert(
                starknet::info::get_tx_info().unbox().chain_id.into() == op.chain_id,
                'wrong chain id'
            );

            assert(starknet::info::get_contract_address() == op.multisig, 'wrong multisig address');

            assert(
                current_expiring_root_and_op_count
                    .valid_until
                    .into() >= starknet::info::get_block_timestamp(),
                'root has expired'
            );

            assert(op.nonce == current_expiring_root_and_op_count.op_count, 'wrong nonce');

            // verify op exists in merkle tree
            let mut encoded_leaf: Bytes = BytesTrait::new_empty()
                .encode(MANY_CHAIN_MULTI_SIG_DOMAIN_SEPARATOR_OP)
                .encode(op.chain_id)
                .encode(op.multisig)
                .encode(op.nonce)
                .encode(op.to)
                .encode(op.selector);
            // encode the data field by looping through
            let mut i = 0;
            while i < op.data.len() {
                encoded_leaf = encoded_leaf.encode(*op.data.at(i));
                i += 1;
            };
            let hashed_leaf = encoded_leaf.keccak();

            assert(
                verify_merkle_proof(proof, current_expiring_root_and_op_count.root, hashed_leaf),
                'proof verification failed'
            );

            let mut new_expiring_root_and_op_count = current_expiring_root_and_op_count;
            new_expiring_root_and_op_count.op_count += 1;

            self.s_expiring_root_and_op_count.write(new_expiring_root_and_op_count);
            // todo: execute
            self._execute(op.to, op.selector, op.data);

            self
                .emit(
                    Event::OpExecuted(
                        OpExecuted {
                            nonce: op.nonce, to: op.to, selector: op.selector, data: op.data
                        }
                    )
                );
        }

        // todo: make onlyOwner
        fn set_config(
            ref self: ContractState,
            signer_addresses: Span<EthAddress>,
            signer_groups: Span<u8>,
            group_quorums: Span<u8>,
            group_parents: Span<u8>,
            clear_root: bool
        ) {
            assert(
                signer_addresses.len() != 0 && signer_addresses.len() <= MAX_NUM_SIGNERS.into(),
                'out of bound signers len'
            );

            assert(signer_addresses.len() == signer_groups.len(), 'signer groups len mismatch');

            assert(
                group_quorums.len() == NUM_GROUPS.into()
                    && group_quorums.len() == group_parents.len(),
                'group quorums/parents mismatch'
            );
        // let mut group_children_counts = 
        }
    }

    #[generate_trait]
    impl InternalFunctions of InternalFunctionsTrait {
        fn _execute(
            ref self: ContractState, target: ContractAddress, selector: felt252, data: Span<felt252>
        ) {
            let _response = call_contract_syscall(target, selector, data).unwrap_syscall();
        }
    }
}

