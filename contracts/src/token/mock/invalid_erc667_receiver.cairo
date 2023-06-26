#[contract]
mod InvalidReceiver {
    use starknet::ContractAddress;
    use array::ArrayTrait;


    struct Storage {
        _supports: bool
    }

    #[constructor]
    fn constructor() {}

    // toggle whether or not receiver says it supports the interface id
    #[external]
    fn set_supports(support: bool) {
        _supports::write(support);
    }


    #[external]
    fn supports_interface(interface_id: u32) -> bool {
        _supports::read()
    }
}
