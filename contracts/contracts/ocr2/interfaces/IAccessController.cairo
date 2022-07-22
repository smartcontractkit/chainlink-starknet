%lang starknet

@contract_interface
namespace IAccessController:
    func has_access(address : felt, data_len : felt, data : felt*) -> (bool : felt):
    end

    func check_access(address : felt):
    end
end
