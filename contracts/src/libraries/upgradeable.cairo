
mod Upgradeable {
    use zeroable::Zeroable;

    use starknet::syscalls::replace_class_syscall;
    use starknet::class_hash::ClassHash;
    use starknet::class_hash::ClassHashZeroable;

    use chainlink::libraries::ownable::Ownable;

    #[event]
    fn Upgraded(implementation: ClassHash) {}

    fn upgrade_only_owner(impl_hash: ClassHash) {
        Ownable::assert_only_owner();
        upgrade(impl_hash);
    }

    fn upgrade(impl_hash: ClassHash) {
        assert(!impl_hash.is_zero(), 'Class hash cannot be zero');
        replace_class_syscall(impl_hash);
        Upgraded(impl_hash);
    }
}
