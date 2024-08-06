use starknet::ContractAddress;
use starknet::{
    EthAddress,
    secp256_trait::{
        Secp256Trait, Secp256PointTrait, recover_public_key, is_signature_entry_valid, Signature
    },
    secp256k1::Secp256k1Point, SyscallResult, SyscallResultTrait
};

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
        signaures: Span<Signature>
    );
    fn execute(ref self: TContractState, op: Op, proof: Span<u256>);
    // todo: check length of group_quorums and group_parents
    fn set_config(
        ref self: TContractState,
        signer_addresses: Span<EthAddress>,
        signer_groups: Span<u8>,
        group_quorums: Span<u8>,
        group_parents: Span<u8>,
        clear_root: bool
    );
    fn get_config(self: @TContractState) -> Config;
    fn get_op_count(self: @TContractState) -> u64;
    fn get_root(self: @TContractState) -> (u256, u32);
    fn get_root_metadata(self: @TContractState) -> RootMetadata;
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
    pre_opcount: u64,
    post_opcount: u64,
    override_previous_root: bool
}

#[derive(Copy, Drop, Serde, starknet::Store)]
struct Op {
    chain_id: u256,
    multisig: ContractAddress,
    nonce: u64,
    to: ContractAddress,
    data: ByteArray
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


#[starknet::contract]
mod ManyChainMultiSig {
    use super::{ExpiringRootAndOpCount, Config, Signer, RootMetadata, Op};
    use starknet::{EthAddress};
    const NUM_GROUPS: u8 = 32;
    const MAX_NUM_SIGNERS: u8 = 200;

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
}

