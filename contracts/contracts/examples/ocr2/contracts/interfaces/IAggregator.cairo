%lang starknet

struct Round:
    member round_id: felt
    member answer: felt
    member block_num: felt
    member observation_timestamp: felt
    member transmission_timestamp: felt
end

@contract_interface
namespace IAggregator:
    func latest_round_data() -> (round: Round):
    end
    
    func decimals() -> (decimals: felt):
    end
end
