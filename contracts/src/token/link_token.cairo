use starknet::ContractAddress;

#[starknet::interface]
trait IMintableToken<TContractState> {
    #[external(v0)]
    fn permissionedMint(ref self: TContractState, account: ContractAddress, amount: u256);
    #[external(v0)]
    fn permissionedBurn(ref self: TContractState, account: ContractAddress, amount: u256);
}

#[starknet::contract]
mod LinkToken {
    use super::IMintableToken;

    use zeroable::Zeroable;

    use openzeppelin::token::erc20::interface::{IERC20, IERC20Dispatcher, IERC20DispatcherTrait};
    use chainlink::libraries::token::erc677::ERC677;
    use chainlink::libraries::ownable::{Ownable, IOwnable};
    use chainlink::libraries::upgradeable::{Upgradeable, IUpgradeable};

    use openzeppelin::token::erc20::ERC20Component;
    use starknet::ContractAddress;
    use starknet::class_hash::ClassHash;

    component!(path: ERC20Component, storage: erc20, event: ERC20Event);

    #[abi(embed_v0)]
    impl ERC20Impl = ERC20Component::ERC20Impl<ContractState>;
    #[abi(embed_v0)]
    impl ERC20MetadataImpl = ERC20Component::ERC20MetadataImpl<ContractState>;
    impl ERC20InternalImpl = ERC20Component::InternalImpl<ContractState>;

    const NAME: felt252 = 'ChainLink Token';
    const SYMBOL: felt252 = 'LINK';

    #[storage]
    struct Storage {
        _minter: ContractAddress,

        #[substorage(v0)]
        erc20: ERC20Component::Storage
    }

    #[event]
    #[derive(Drop, starknet::Event)]
    enum Event {
        #[flat]
        ERC20Event: ERC20Component::Event
    }

    //
    // IMintableToken (StarkGate)
    //
    #[external(v0)]
    impl MintableToken of IMintableToken<ContractState> {
        fn permissionedMint(ref self: ContractState, account: ContractAddress, amount: u256) {
            only_minter(@self);
            self.erc20._mint(account, amount);
        }

        fn permissionedBurn(ref self: ContractState, account: ContractAddress, amount: u256) {
            only_minter(@self);
            self.erc20._burn(account, amount);
        }
    }


    #[constructor]
    fn constructor(ref self: ContractState, minter: ContractAddress, owner: ContractAddress) {
        self.erc20.initializer(NAME, SYMBOL);
        assert(!minter.is_zero(), 'minter is 0');
        self._minter.write(minter);
        let mut ownable = Ownable::unsafe_new_contract_state();
        Ownable::constructor(ref ownable, owner); // Ownable::initializer(owner);
    }

    // TODO #[view]
    fn minter(self: @ContractState) -> ContractAddress {
        self._minter.read()
    }

    // TODO #[view]
    fn type_and_version(self: @ContractState) -> felt252 {
        'LinkToken 1.0.0'
    }

    // 
    // ERC677
    //

    // TODO:
    // #[external(v0)]
    // fn transfer_and_call(
    //     ref self: ContractState, to: ContractAddress, value: u256, data: Array<felt252>
    // ) -> bool {
    //     let mut erc677 = ERC677::unsafe_new_contract_state();
    //     ERC677::transfer_and_call(ref erc677, to, value, data)
    // }

    //
    //  Upgradeable
    //
    #[external(v0)]
    impl UpgradeableImpl of IUpgradeable<ContractState> {
        fn upgrade(ref self: ContractState, new_impl: ClassHash) {
            let ownable = Ownable::unsafe_new_contract_state();
            Ownable::assert_only_owner(@ownable);
            Upgradeable::upgrade(new_impl)
        }
    }

    //
    // Ownership
    //

    #[external(v0)]
    impl OwnableImpl of IOwnable<ContractState> {
        fn owner(self: @ContractState) -> ContractAddress {
            let state = Ownable::unsafe_new_contract_state();
            Ownable::OwnableImpl::owner(@state)
        }

        fn proposed_owner(self: @ContractState) -> ContractAddress {
            let state = Ownable::unsafe_new_contract_state();
            Ownable::OwnableImpl::proposed_owner(@state)
        }

        fn transfer_ownership(ref self: ContractState, new_owner: ContractAddress) {
            let mut state = Ownable::unsafe_new_contract_state();
            Ownable::OwnableImpl::transfer_ownership(ref state, new_owner)
        }

        fn accept_ownership(ref self: ContractState) {
            let mut state = Ownable::unsafe_new_contract_state();
            Ownable::OwnableImpl::accept_ownership(ref state)
        }

        fn renounce_ownership(ref self: ContractState) {
            let mut state = Ownable::unsafe_new_contract_state();
            Ownable::OwnableImpl::renounce_ownership(ref state)
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
