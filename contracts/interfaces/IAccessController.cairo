%lang starknet

@contract_interface
namespace IAccessController:
    func has_access(address: felt) -> (bool: felt):
    end

    func check_access(address: felt):
    end
end