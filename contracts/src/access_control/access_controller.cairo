#[starknet::contract]
mod AccessController {
    use starknet::ContractAddress;
    use starknet::class_hash::ClassHash;

    use openzeppelin::access::ownable::OwnableComponent;
    use openzeppelin::upgrades::UpgradeableComponent;

    use chainlink::libraries::access_control::{AccessControlComponent, IAccessController};
    use chainlink::libraries::type_and_version::ITypeAndVersion;
    use chainlink::libraries::upgrades::v2::owner_upgradeable::OwnerUpgradeableComponent;

    component!(path: OwnableComponent, storage: ownable, event: OwnableEvent);
    component!(path: AccessControlComponent, storage: access_control, event: AccessControlEvent);
    component!(path: UpgradeableComponent, storage: upgradeable, event: UpgradeableEvent);
    component!(
        path: OwnerUpgradeableComponent, storage: owner_upgradeable, event: OwnerUpgradeableEvent,
    );

    #[abi(embed_v0)]
    impl OwnableImpl = OwnableComponent::OwnableTwoStepImpl<ContractState>;
    impl OwnableInternalImpl = OwnableComponent::InternalImpl<ContractState>;

    #[abi(embed_v0)]
    impl AccessControlImpl =
        AccessControlComponent::AccessControlImpl<ContractState>;
    impl AccessControlInternalImpl = AccessControlComponent::InternalImpl<ContractState>;

    impl UpgradeableInternalImpl = UpgradeableComponent::InternalImpl<ContractState>;

    #[abi(embed_v0)]
    impl OwnerUpgradeableImpl =
        OwnerUpgradeableComponent::OwnerUpgradeableImpl<ContractState>;

    #[event]
    #[derive(Drop, starknet::Event)]
    enum Event {
        #[flat]
        OwnableEvent: OwnableComponent::Event,
        #[flat]
        AccessControlEvent: AccessControlComponent::Event,
        #[flat]
        UpgradeableEvent: UpgradeableComponent::Event,
        #[flat]
        OwnerUpgradeableEvent: OwnerUpgradeableComponent::Event,
    }

    #[storage]
    struct Storage {
        #[substorage(v0)]
        ownable: OwnableComponent::Storage,
        #[substorage(v0)]
        access_control: AccessControlComponent::Storage,
        #[substorage(v0)]
        upgradeable: UpgradeableComponent::Storage,
        #[substorage(v0)]
        owner_upgradeable: OwnerUpgradeableComponent::Storage,
    }

    #[constructor]
    fn constructor(ref self: ContractState, owner_address: ContractAddress) {
        self.ownable.initializer(owner_address);
        self.access_control.initializer(true);
    }

    #[abi(embed_v0)]
    impl TypeAndVersionImpl of ITypeAndVersion<ContractState> {
        fn type_and_version(self: @ContractState) -> felt252 {
            'AccessController 1.0.0'
        }
    }
}
