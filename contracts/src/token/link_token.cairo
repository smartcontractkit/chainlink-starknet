use starknet::ContractAddress;

// https://github.com/starknet-io/starkgate-contracts/blob/v2.0/src/cairo/mintable_token_interface.cairo
#[starknet::interface]
trait IMintableToken<TContractState> {
    fn permissioned_mint(ref self: TContractState, account: ContractAddress, amount: u256);
    fn permissioned_burn(ref self: TContractState, account: ContractAddress, amount: u256);
}

#[starknet::contract]
mod LinkToken {
    use starknet::ContractAddress;
    use starknet::class_hash::ClassHash;

    use zeroable::Zeroable;

    use openzeppelin::token::erc20::ERC20Component;
    use openzeppelin::access::ownable::OwnableComponent;

    use super::IMintableToken;
    use openzeppelin::token::erc20::interface::{IERC20, IERC20Dispatcher, IERC20DispatcherTrait};
    use chainlink::libraries::token::erc677::ERC677Component;
    use chainlink::libraries::type_and_version::ITypeAndVersion;
    use chainlink::libraries::upgradeable::{Upgradeable, IUpgradeable};

    component!(path: OwnableComponent, storage: ownable, event: OwnableEvent);
    component!(path: ERC20Component, storage: erc20, event: ERC20Event);
    component!(path: ERC677Component, storage: erc677, event: ERC677Event);

    #[abi(embed_v0)]
    impl OwnableImpl = OwnableComponent::OwnableTwoStepImpl<ContractState>;
    impl InternalImpl = OwnableComponent::InternalImpl<ContractState>;

    #[abi(embed_v0)]
    impl ERC20Impl = ERC20Component::ERC20Impl<ContractState>;
    #[abi(embed_v0)]
    impl ERC20MetadataImpl = ERC20Component::ERC20MetadataImpl<ContractState>;
    impl ERC20InternalImpl = ERC20Component::InternalImpl<ContractState>;

    #[abi(embed_v0)]
    impl ERC677Impl = ERC677Component::ERC677Impl<ContractState>;

    #[storage]
    struct Storage {
        #[substorage(v0)]
        ownable: OwnableComponent::Storage,
        _minter: ContractAddress,
        #[substorage(v0)]
        erc20: ERC20Component::Storage,
        #[substorage(v0)]
        erc677: ERC677Component::Storage
    }

    #[event]
    #[derive(Drop, starknet::Event)]
    enum Event {
        #[flat]
        OwnableEvent: OwnableComponent::Event,
        #[flat]
        ERC20Event: ERC20Component::Event,
        #[flat]
        ERC677Event: ERC677Component::Event
    }

    //
    // IMintableToken (StarkGate)
    //
    #[abi(embed_v0)]
    impl MintableToken of IMintableToken<ContractState> {
        fn permissioned_mint(ref self: ContractState, account: ContractAddress, amount: u256) {
            only_minter(@self);
            self.erc20._mint(account, amount);
        }

        fn permissioned_burn(ref self: ContractState, account: ContractAddress, amount: u256) {
            only_minter(@self);
            self.erc20._burn(account, amount);
        }
    }


    #[constructor]
    fn constructor(ref self: ContractState, minter: ContractAddress, owner: ContractAddress) {
        let name = "ChainLink Token";
        let symbol = "LINK";
        self.erc20.initializer(name, symbol);
        assert(!minter.is_zero(), 'minter is 0');
        self._minter.write(minter);
        self.ownable.initializer(owner);
    }

    // TODO #[view]
    fn minter(self: @ContractState) -> ContractAddress {
        self._minter.read()
    }

    #[abi(embed_v0)]
    impl TypeAndVersionImpl of ITypeAndVersion<ContractState> {
        fn type_and_version(self: @ContractState) -> felt252 {
            'LinkToken 1.0.0'
        }
    }

    #[abi(embed_v0)]
    impl UpgradeableImpl of IUpgradeable<ContractState> {
        fn upgrade(ref self: ContractState, new_impl: ClassHash) {
            self.ownable.assert_only_owner();
            Upgradeable::upgrade(new_impl)
        }
    }

    // fn increase_allowance(ref self: ContractState, spender: ContractAddress, added_value: u256) -> bool {
    //     let mut state = ERC20::unsafe_new_contract_state();
    //     ERC20::ERC20Impl::increase_allowance(ref state, spender, added_value)
    // }

    // fn decrease_allowance(ref self: ContractState, spender: ContractAddress, subtracted_value: u256) -> bool {
    //     let mut state = ERC20::unsafe_new_contract_state();
    //     ERC20::ERC20Impl::decrease_allowance(ref state, spender, subtracted_value)
    // }

    //
    // Internal
    //

    fn only_minter(self: @ContractState) {
        let caller = starknet::get_caller_address();
        let minter = self._minter.read();
        assert(caller == minter, 'only minter');
    }
}
