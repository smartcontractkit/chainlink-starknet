use starknet::class_hash::ClassHash;

#[contract]
mod MockUpgradeable {
    use starknet::class_hash::ClassHash;

    use chainlink::libraries::upgradeable::Upgradeable;

    #[constructor]
    fn constructor() {}

    #[view]
    fn foo() -> bool {
        true
    }

    #[external]
    fn upgrade(new_impl: ClassHash) {
        Upgradeable::upgrade(new_impl)
    }
}
