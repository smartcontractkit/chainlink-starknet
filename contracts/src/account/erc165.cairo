// copied from https://github.com/OpenZeppelin/cairo-contracts/pull/583

const IERC165_ID: u32 = 0x01ffc9a7_u32;
const INVALID_ID: u32 = 0xffffffff_u32;

trait IERC165 {
    fn supports_interface(interface_id: u32) -> bool;
}

#[contract]
mod ERC165 {
    use super::IERC165_ID;
    use super::INVALID_ID;
    use super::IERC165;

    struct Storage {
        supported_interfaces: LegacyMap<u32, bool>, 
    }

    impl ERC165 of IERC165 {
        fn supports_interface(interface_id: u32) -> bool {
            if interface_id == IERC165_ID {
                return true;
            }
            supported_interfaces::read(interface_id)
        }
    }

    #[view]
    fn supports_interface(interface_id: u32) -> bool {
        ERC165::supports_interface(interface_id)
    }

    #[internal]
    fn register_interface(interface_id: u32) {
        assert(interface_id != INVALID_ID, 'Invalid id');
        supported_interfaces::write(interface_id, true);
    }

    #[internal]
    fn deregister_interface(interface_id: u32) {
        assert(interface_id != IERC165_ID, 'Invalid id');
        supported_interfaces::write(interface_id, false);
    }
}
