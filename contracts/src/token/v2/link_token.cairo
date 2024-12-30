use starknet::ContractAddress;

// This token is deployed by the StarkGate bridge

// https://github.com/starknet-io/starkgate-contracts/blob/eedee8304e8c407c2e0e03c83187dbc5dcc6787e/src/cairo/mintable_token_interface.cairo
#[starknet::interface]
trait IMintableToken<TContractState> {
    fn permissioned_mint(ref self: TContractState, account: ContractAddress, amount: u256);
    fn permissioned_burn(ref self: TContractState, account: ContractAddress, amount: u256);
}

// allows setting and getting the minter
#[starknet::interface]
trait IMinter<TContractState> {
    fn set_minter(ref self: TContractState, new_minter: ContractAddress);
    fn minter(self: @TContractState) -> ContractAddress;
}

#[starknet::contract]
mod LinkToken {
    use starknet::{contract_address_const, ContractAddress, class_hash::ClassHash};
    use zeroable::Zeroable;
    use openzeppelin::{
        token::erc20::{
            ERC20Component, interface::{IERC20, IERC20Dispatcher, IERC20DispatcherTrait},
        },
        access::ownable::OwnableComponent, upgrades::UpgradeableComponent,
    };
    use super::{IMintableToken, IMinter};
    use chainlink::libraries::{
        token::v2::erc677::ERC677Component, type_and_version::ITypeAndVersion,
        upgrades::v1::upgradeable::{Upgradeable, IUpgradeable},
        upgrades::v2::owner_upgradeable::OwnerUpgradeableComponent,
    };

    component!(path: OwnableComponent, storage: ownable, event: OwnableEvent);
    component!(path: ERC20Component, storage: erc20, event: ERC20Event);
    component!(path: ERC677Component, storage: erc677, event: ERC677Event);
    component!(path: UpgradeableComponent, storage: upgradeable, event: UpgradeableEvent);
    component!(
        path: OwnerUpgradeableComponent, storage: owner_upgradeable, event: OwnerUpgradeableEvent,
    );

    #[abi(embed_v0)]
    impl OwnableImpl = OwnableComponent::OwnableTwoStepImpl<ContractState>;
    impl OwnableInternalImpl = OwnableComponent::InternalImpl<ContractState>;

    #[abi(embed_v0)]
    impl ERC20Impl = ERC20Component::ERC20Impl<ContractState>;
    #[abi(embed_v0)]
    impl ERC20MetadataImpl = ERC20Component::ERC20MetadataImpl<ContractState>;
    impl ERC20InternalImpl = ERC20Component::InternalImpl<ContractState>;

    #[abi(embed_v0)]
    impl ERC677Impl = ERC677Component::ERC677Impl<ContractState>;

    impl UpgradeableInternalImpl = UpgradeableComponent::InternalImpl<ContractState>;

    #[abi(embed_v0)]
    impl OwnerUpgradeableImpl =
        OwnerUpgradeableComponent::OwnerUpgradeableImpl<ContractState>;

    #[storage]
    struct Storage {
        LinkTokenV2_minter: ContractAddress,
        #[substorage(v0)]
        ownable: OwnableComponent::Storage,
        #[substorage(v0)]
        erc20: ERC20Component::Storage,
        #[substorage(v0)]
        erc677: ERC677Component::Storage,
        #[substorage(v0)]
        upgradeable: UpgradeableComponent::Storage,
        #[substorage(v0)]
        owner_upgradeable: OwnerUpgradeableComponent::Storage,
    }

    #[derive(Drop, starknet::Event)]
    struct LinkTokenV2NewMinter {
        old_minter: ContractAddress,
        new_minter: ContractAddress,
    }

    #[event]
    #[derive(Drop, starknet::Event)]
    enum Event {
        LinkTokenV2NewMinter: LinkTokenV2NewMinter,
        #[flat]
        OwnableEvent: OwnableComponent::Event,
        #[flat]
        ERC20Event: ERC20Component::Event,
        #[flat]
        ERC677Event: ERC677Component::Event,
        #[flat]
        UpgradeableEvent: UpgradeableComponent::Event,
        #[flat]
        OwnerUpgradeableEvent: OwnerUpgradeableComponent::Event,
    }

    #[constructor]
    fn constructor(
        ref self: ContractState,
        _name_ignore: felt252,
        _symbol_ignore: felt252,
        _decimals_ignore: u8,
        _initial_supply_ignore: u256,
        _initial_recipient_ignore: ContractAddress,
        initial_minter: ContractAddress,
        owner: ContractAddress,
        _upgrade_delay_ignore: u64,
    ) {
        let name = "ChainLink Token";
        let symbol = "LINK";

        self.erc20.initializer(name, symbol);
        self.ownable.initializer(owner);

        assert(!initial_minter.is_zero(), 'minter is 0');
        self.LinkTokenV2_minter.write(initial_minter);

        self
            .emit(
                Event::LinkTokenV2NewMinter(
                    LinkTokenV2NewMinter {
                        old_minter: contract_address_const::<0>(), new_minter: initial_minter,
                    },
                ),
            );
    }

    impl HooksImpl of ERC20Component::ERC20HooksTrait<ContractState> {
        fn before_update(
            ref self: ERC20Component::ComponentState::<ContractState>,
            from: ContractAddress,
            recipient: ContractAddress,
            amount: u256,
        ) {}

        fn after_update(
            ref self: ERC20Component::ComponentState::<ContractState>,
            from: ContractAddress,
            recipient: ContractAddress,
            amount: u256,
        ) {}
    }

    #[abi(embed_v0)]
    impl MintableToken of IMintableToken<ContractState> {
        fn permissioned_mint(ref self: ContractState, account: ContractAddress, amount: u256) {
            self._only_minter();
            self.erc20.mint(account, amount);
        }

        fn permissioned_burn(ref self: ContractState, account: ContractAddress, amount: u256) {
            self._only_minter();
            self.erc20.burn(account, amount);
        }
    }

    #[abi(embed_v0)]
    impl Minter of IMinter<ContractState> {
        fn set_minter(ref self: ContractState, new_minter: ContractAddress) {
            self.ownable.assert_only_owner();

            let prev_minter = self.LinkTokenV2_minter.read();
            assert(new_minter != prev_minter, 'is minter already');
            assert(!new_minter.is_zero(), 'minter is 0');

            self.LinkTokenV2_minter.write(new_minter);

            self
                .emit(
                    Event::LinkTokenV2NewMinter(
                        LinkTokenV2NewMinter { old_minter: prev_minter, new_minter: new_minter },
                    ),
                );
        }

        fn minter(self: @ContractState) -> ContractAddress {
            self.LinkTokenV2_minter.read()
        }
    }

    #[abi(embed_v0)]
    impl TypeAndVersionImpl of ITypeAndVersion<ContractState> {
        fn type_and_version(self: @ContractState) -> felt252 {
            'LinkToken 2.0.0'
        }
    }

    //
    // Internal
    //

    #[generate_trait]
    impl InternalFunctions of InternalFunctionsTrait {
        fn _only_minter(self: @ContractState) {
            let caller = starknet::get_caller_address();
            let minter = self.LinkTokenV2_minter.read();
            assert(caller == minter, 'only minter');
        }
    }
}
