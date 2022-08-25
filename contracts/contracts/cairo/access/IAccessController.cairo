%lang starknet

@contract_interface
namespace IAccessController:
    func has_access(user : felt, data_len : felt, data : felt*) -> (bool : felt):
    end

    func check_access(user : felt):
    end
end
