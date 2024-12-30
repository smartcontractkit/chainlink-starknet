use starknet::class_hash::ClassHash;

#[starknet::interface]
trait IFoo<TContractState> {
    fn foo(self: @TContractState) -> bool;
}

#[starknet::contract]
mod MockOwnerUpgradeable {
    use starknet::class_hash::ClassHash;
    use starknet::ContractAddress;

    use openzeppelin::access::ownable::OwnableComponent;
    use openzeppelin::upgrades::UpgradeableComponent;

    use chainlink::libraries::upgrades::v2::owner_upgradeable::OwnerUpgradeableComponent;

    component!(path: OwnableComponent, storage: ownable, event: OwnableEvent);
    component!(path: UpgradeableComponent, storage: upgradeable, event: UpgradeableEvent);
    component!(
        path: OwnerUpgradeableComponent, storage: owner_upgradeable, event: OwnerUpgradeableEvent,
    );

    #[abi(embed_v0)]
    impl OwnableImpl = OwnableComponent::OwnableTwoStepImpl<ContractState>;
    impl OwnableInternalImpl = OwnableComponent::InternalImpl<ContractState>;

    impl UpgradeableInternalImpl = UpgradeableComponent::InternalImpl<ContractState>;

    #[abi(embed_v0)]
    impl OwnerUpgradeableImpl =
        OwnerUpgradeableComponent::OwnerUpgradeableImpl<ContractState>;

    #[storage]
    struct Storage {
        #[substorage(v0)]
        ownable: OwnableComponent::Storage,
        #[substorage(v0)]
        upgradeable: UpgradeableComponent::Storage,
        #[substorage(v0)]
        owner_upgradeable: OwnerUpgradeableComponent::Storage,
    }

    #[event]
    #[derive(Drop, starknet::Event)]
    enum Event {
        #[flat]
        OwnableEvent: OwnableComponent::Event,
        #[flat]
        UpgradeableEvent: UpgradeableComponent::Event,
        #[flat]
        OwnerUpgradeableEvent: OwnerUpgradeableComponent::Event,
    }

    #[constructor]
    fn constructor(ref self: ContractState, owner: ContractAddress) {
        self.ownable.initializer(owner);
    }

    #[abi(embed_v0)]
    impl FooImpl of super::IFoo<ContractState> {
        fn foo(self: @ContractState) -> bool {
            true
        }
    }
}
