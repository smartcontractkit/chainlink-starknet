use starknet::class_hash::ClassHash;

// TODO: drop for OZ upgradeable

#[starknet::interface]
trait IUpgradeable<TContractState> {
    // note: any contract that uses this module will have a mutable reference to contract state
    fn upgrade(ref self: TContractState, new_impl: ClassHash);
}

#[derive(Drop, starknet::Event)]
struct Upgraded {
    #[key]
    new_impl: ClassHash
}

mod Upgradeable {
    use zeroable::Zeroable;

    use starknet::SyscallResult;
    use starknet::SyscallResultTrait;
    use starknet::syscalls::replace_class_syscall;
    use starknet::class_hash::ClassHash;
    use starknet::class_hash::ClassHashZeroable;

    // this method assumes replace_class_syscall has a very low possibility of being deprecated
    // but if it does, we will either have upgraded the contract to be non-upgradeable by then
    // because the starknet ecosystem has stabilized or we will be able to upgrade the contract to the proxy pattern
    // #[internal]
    fn upgrade(new_impl: ClassHash) {
        assert(!new_impl.is_zero(), 'Class hash cannot be zero');
        replace_class_syscall(new_impl).unwrap_syscall();
    // TODO: Upgraded(new_impl);
    }
}
