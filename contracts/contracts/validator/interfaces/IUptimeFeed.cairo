%lang starknet

from ocr2.interfaces.IAggregator import Round

@contract_interface
namespace IUptimeFeed:
    func latest_round_data() -> (round : Round):
    end
end
