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

    use starknet::ContractAddress;
    use starknet::class_hash::ClassHash;

    use openzeppelin::token::erc20::ERC20;
    use openzeppelin::token::erc20::interface::{IERC20, IERC20Dispatcher, IERC20DispatcherTrait};
    use chainlink::libraries::token::erc677::ERC677;
    use chainlink::libraries::ownable::{Ownable, IOwnable};
    use chainlink::libraries::upgradeable::{Upgradeable, IUpgradeable};

    const NAME: felt252 = 'ChainLink Token';
    const SYMBOL: felt252 = 'LINK';

    #[storage]
    struct Storage {
        _minter: ContractAddress,
    }

    //
    // IMintableToken (StarkGate)
    //
    #[external(v0)]
    impl MintableToken of IMintableToken<ContractState> {
        fn permissionedMint(ref self: ContractState, account: ContractAddress, amount: u256) {
            only_minter(@self);
            let mut state = ERC20::unsafe_new_contract_state();
            ERC20::InternalImpl::_mint(ref state, account, amount);
        }

        fn permissionedBurn(ref self: ContractState, account: ContractAddress, amount: u256) {
            only_minter(@self);
            let mut state = ERC20::unsafe_new_contract_state();
            ERC20::InternalImpl::_burn(ref state, account, amount);
        }
    }


    #[constructor]
    fn constructor(ref self: ContractState, minter: ContractAddress, owner: ContractAddress) {
        let mut state = ERC20::unsafe_new_contract_state();
        ERC20::InternalImpl::initializer(ref state, NAME, SYMBOL);
        assert(!minter.is_zero(), 'minter is 0');
        self._minter.write(minter);
        let mut ownable = Ownable::unsafe_new_contract_state();
        Ownable::constructor(ref ownable, owner); // Ownable::initializer(owner);
    }

    #[view]
    fn minter(self: @ContractState) -> ContractAddress {
        self._minter.read()
    }

    #[view]
    fn type_and_version(self: @ContractState) -> felt252 {
        'LinkToken 1.0.0'
    }

    // 
    // ERC677
    //

    #[external(v0)]
    fn transfer_and_call(
        ref self: ContractState, to: ContractAddress, value: u256, data: Array<felt252>
    ) -> bool {
        let mut erc677 = ERC677::unsafe_new_contract_state();
        ERC677::transfer_and_call(ref erc677, to, value, data)
    }

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

    //
    // ERC20
    //

    #[external(v0)]
    impl ERC20Impl of IERC20<ContractState> {
        fn name(self: @ContractState) -> felt252 {
            let state = ERC20::unsafe_new_contract_state();
            ERC20::ERC20Impl::name(@state)
        }

        fn symbol(self: @ContractState) -> felt252 {
            let state = ERC20::unsafe_new_contract_state();
            ERC20::ERC20Impl::symbol(@state)
        }

        fn decimals(self: @ContractState) -> u8 {
            let state = ERC20::unsafe_new_contract_state();
            ERC20::ERC20Impl::decimals(@state)
        }

        fn total_supply(self: @ContractState) -> u256 {
            let state = ERC20::unsafe_new_contract_state();
            ERC20::ERC20Impl::total_supply(@state)
        }

        fn balance_of(self: @ContractState, account: ContractAddress) -> u256 {
            let state = ERC20::unsafe_new_contract_state();
            ERC20::ERC20Impl::balance_of(@state, account)
        }

        fn allowance(
            self: @ContractState, owner: ContractAddress, spender: ContractAddress
        ) -> u256 {
            let state = ERC20::unsafe_new_contract_state();
            ERC20::ERC20Impl::allowance(@state, owner, spender)
        }

        fn transfer(ref self: ContractState, recipient: ContractAddress, amount: u256) -> bool {
            let mut state = ERC20::unsafe_new_contract_state();
            ERC20::ERC20Impl::transfer(ref state, recipient, amount)
        }

        fn transfer_from(
            ref self: ContractState,
            sender: ContractAddress,
            recipient: ContractAddress,
            amount: u256
        ) -> bool {
            let mut state = ERC20::unsafe_new_contract_state();
            ERC20::ERC20Impl::transfer_from(ref state, sender, recipient, amount)
        }

        fn approve(ref self: ContractState, spender: ContractAddress, amount: u256) -> bool {
            let mut state = ERC20::unsafe_new_contract_state();
            ERC20::ERC20Impl::approve(ref state, spender, amount)
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
