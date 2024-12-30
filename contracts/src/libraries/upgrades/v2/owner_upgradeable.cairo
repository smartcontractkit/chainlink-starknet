#[starknet::component]
mod OwnerUpgradeableComponent {
    use openzeppelin::{
        access::ownable::{
            OwnableComponent, OwnableComponent::InternalTrait as OwnableInternalTrait,
        },
        upgrades::{
            upgradeable::{
                UpgradeableComponent,
                UpgradeableComponent::InternalTrait as UpgradeableInternalTrait,
            },
            interface::IUpgradeable,
        },
    };
    use starknet::class_hash::ClassHash;

    #[storage]
    struct Storage {}

    #[event]
    #[derive(Drop, starknet::Event)]
    enum Event {}

    #[embeddable_as(OwnerUpgradeableImpl)]
    pub impl OwnerUpgradeable<
        TContractState,
        +HasComponent<TContractState>,
        impl Ownable: OwnableComponent::HasComponent<TContractState>,
        impl Upgradeable: UpgradeableComponent::HasComponent<TContractState>,
        +Drop<TContractState>,
    > of IUpgradeable<ComponentState<TContractState>> {
        fn upgrade(ref self: ComponentState<TContractState>, new_class_hash: ClassHash) {
            let mut ownable_component = get_dep_component_mut!(ref self, Ownable);
            ownable_component.assert_only_owner();

            let mut upgradeable_component = get_dep_component_mut!(ref self, Upgradeable);
            upgradeable_component.upgrade(new_class_hash);
        }
    }
}

