%lang starknet

struct RoundFeed:
    member status : felt
    member started_at : felt
    member updated_at : felt
end

@contract_interface
namespace IUptimeFeed:
    func latest_round_data() -> (round : RoundFeed):
    end
end