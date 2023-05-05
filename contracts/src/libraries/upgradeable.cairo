
mod Upgradeable {
    use zeroable::Zeroable;

    use starknet::SyscallResult;
    use starknet::SyscallResultTrait;
    use starknet::syscalls::replace_class_syscall;
    use starknet::class_hash::ClassHash;
    use starknet::class_hash::ClassHashZeroable;

    #[event]
    fn Upgraded(new_class_hash: ClassHash) {}

    // this method assumes replace_class_syscall has a very low possibility of being deprecated
    // but if it does, we will either have upgraded the contract to be non-upgradeable by then
    // because the starknet ecosystem has stabilized or we will be able to upgrade the contract to the proxy pattern
    fn upgrade(new_class_hash: ClassHash) {
        assert(!new_class_hash.is_zero(), 'Class hash cannot be zero');
        replace_class_syscall(new_class_hash).unwrap_syscall();
        Upgraded(new_class_hash);
    }
}
