%lang starknet

@contract_interface
namespace ISequencerUptimeFeed:
    func update_status(status : felt, timestamp : felt):
    end
end
