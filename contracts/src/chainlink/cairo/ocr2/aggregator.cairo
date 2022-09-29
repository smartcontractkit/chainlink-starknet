%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin, SignatureBuiltin, BitwiseBuiltin
from starkware.cairo.common.alloc import alloc
from starkware.cairo.common.registers import get_fp_and_pc
from starkware.cairo.common.hash_state import (
    hash_init,
    hash_finalize,
    hash_update,
    hash_update_single,
)
from starkware.cairo.common.signature import verify_ecdsa_signature
from starkware.cairo.common.bitwise import bitwise_and
from starkware.cairo.common.bool import TRUE
from starkware.cairo.common.math import (
    abs_value,
    split_felt,
    assert_lt_felt,
    assert_lt,
    assert_le,
    assert_not_zero,
    assert_not_equal,
    assert_nn_le,
    assert_nn,
    assert_in_range,
    unsigned_div_rem,
)
from starkware.cairo.common.pow import pow
from starkware.cairo.common.uint256 import Uint256, uint256_sub, uint256_lt, uint256_le

from starkware.starknet.common.syscalls import (
    get_caller_address,
    get_contract_address,
    get_block_timestamp,
    get_block_number,
    get_tx_info,
)

from openzeppelin.utils.constants.library import UINT8_MAX

from openzeppelin.token.erc20.IERC20 import IERC20

from chainlink.cairo.access.IAccessController import IAccessController

from chainlink.cairo.utils import felt_to_uint256, uint256_to_felt

from chainlink.cairo.access.ownable import Ownable

from chainlink.cairo.access.SimpleReadAccessController.library import SimpleReadAccessController

from chainlink.cairo.ocr2.IAggregator import NewTransmission, Round

# ---

const MAX_ORACLES = 31

const GIGA = 10 ** 9

const UINT32_MAX = (2 ** 32) - 1
const INT128_MAX = (2 ** (128 - 1)) - 1

# Maximum number of faulty oracles
@storage_var
func Aggregator_f() -> (f : felt):
end

@storage_var
func Aggregator_latest_epoch_and_round() -> (res : felt):
end

@storage_var
func Aggregator_latest_aggregator_round_id() -> (round_id : felt):
end

using Range = (min : felt, max : felt)

@storage_var
func Aggregator_answer_range() -> (range : Range):
end

@storage_var
func Aggregator_decimals() -> (decimals : felt):
end

@storage_var
func Aggregator_description() -> (description : felt):
end

#

@storage_var
func Aggregator_latest_config_block_number() -> (block : felt):
end

@storage_var
func Aggregator_config_count() -> (count : felt):
end

@storage_var
func Aggregator_latest_config_digest() -> (digest : felt):
end

@storage_var
func Aggregator_oracles_len() -> (len : felt):
end

# TODO: should we pack into (index, payment) = split_felt()? index is u8, payment is u128
struct Oracle:
    member index : felt

    # entire supply of LINK always fits into u96, so felt is safe to use
    member payment_juels : felt
end

@storage_var
func Aggregator_transmitters(pkey : felt) -> (index : Oracle):
end

@storage_var
func Aggregator_signers(pkey : felt) -> (index : felt):
end

@storage_var
func Aggregator_signers_list(index : felt) -> (pkey : felt):
end

@storage_var
func Aggregator_transmitters_list(index : felt) -> (pkey : felt):
end

@storage_var
func reward_from_aggregator_round_id_(index : felt) -> (round_id : felt):
end

# ---

struct Transmission:
    member answer : felt
    member block_num : felt
    member observation_timestamp : felt
    member transmission_timestamp : felt
end

@storage_var
func Aggregator_transmissions(round_id : felt) -> (transmission : Transmission):
end

# ---

@constructor
func constructor{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    owner : felt,
    link : felt,
    min_answer : felt,
    max_answer : felt,
    billing_access_controller : felt,
    decimals : felt,
    description : felt,
):
    Ownable.initializer(owner)
    SimpleReadAccessController.initialize(owner)  # This also calls Ownable.initializer
    Aggregator_link_token.write(link)
    Aggregator_billing_access_controller.write(billing_access_controller)

    assert_lt(min_answer, max_answer)
    let range : Range = (min=min_answer, max=max_answer)
    Aggregator_answer_range.write(range)

    with_attr error_message("decimals exceed 2^8"):
        assert_lt(decimals, UINT8_MAX)
    end
    Aggregator_decimals.write(decimals)
    Aggregator_description.write(description)
    return ()
end

# --- Validation ---

# NOTE: Currently unimplemented:
# - Can't set a gas limit on the validator call
# - Can't catch errors in calls so validation could block submission

# --- Configuration

@event
func ConfigSet(
    previous_config_block_number : felt,
    latest_config_digest : felt,
    config_count : felt,
    oracles_len : felt,
    oracles : OracleConfig*,
    f : felt,
    onchain_config_len : felt,
    onchain_config : felt*,
    offchain_config_version : felt,
    offchain_config_len : felt,
    offchain_config : felt*,
):
end

struct OracleConfig:
    member signer : felt
    member transmitter : felt
end

struct OnchainConfig:
    member version : felt
    member min_answer : felt
    member max_answer : felt
end

@external
func set_config{
    syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, bitwise_ptr : BitwiseBuiltin*, range_check_ptr
}(
    oracles_len : felt,
    oracles : OracleConfig*,
    f : felt,
    onchain_config_len : felt,
    onchain_config : felt*,
    offchain_config_version : felt,
    offchain_config_len : felt,
    offchain_config : felt*,
) -> (digest : felt):
    alloc_locals
    Ownable.assert_only_owner()

    assert_nn_le(oracles_len, MAX_ORACLES)  # oracles_len <= MAX_ORACLES
    assert_lt(3 * f, oracles_len)  # 3 * f < oracles_len
    assert_nn(f)  # f is positive

    # Notice: onchain_config is always zero since we don't allow configuring it yet after deployment.
    # The contract still computes the onchain_config while digesting the config using min/maxAnswer set on construction.
    with_attr error_message("onchain_config must be empty"):
        assert onchain_config_len = 0
    end

    let (answer_range : Range) = Aggregator_answer_range.read()
    local computed_onchain_config : OnchainConfig = OnchainConfig(version=1, min_answer=answer_range.min, max_answer=answer_range.max)
    # cast to felt* and use OnchainConfig.SIZE as len
    let (__fp__, _) = get_fp_and_pc()
    let onchain_config = cast(&computed_onchain_config, felt*)

    # pay out existing oracles
    pay_oracles()

    # remove old signers/transmitters
    let (len) = Aggregator_oracles_len.read()
    remove_oracles(len)

    let (latest_round_id) = Aggregator_latest_aggregator_round_id.read()

    # add new oracles (also sets oracle_len_)
    add_oracles(oracles, 0, oracles_len, latest_round_id)

    Aggregator_f.write(f)
    let (block_num : felt) = get_block_number()
    let (prev_block_num) = Aggregator_latest_config_block_number.read()
    Aggregator_latest_config_block_number.write(block_num)
    # update config count
    let (config_count) = Aggregator_config_count.read()
    let config_count = config_count + 1
    Aggregator_config_count.write(config_count)
    # calculate and store config digest
    let (contract_address) = get_contract_address()
    let (tx_info) = get_tx_info()
    let (digest) = config_digest_from_data(
        tx_info.chain_id,
        contract_address,
        config_count,
        oracles_len,
        oracles,
        f,
        OnchainConfig.SIZE,
        onchain_config,
        offchain_config_version,
        offchain_config_len,
        offchain_config,
    )
    Aggregator_latest_config_digest.write(digest)

    # reset epoch & round
    Aggregator_latest_epoch_and_round.write(0)

    ConfigSet.emit(
        previous_config_block_number=prev_block_num,
        latest_config_digest=digest,
        config_count=config_count,
        oracles_len=oracles_len,
        oracles=oracles,
        f=f,
        onchain_config_len=OnchainConfig.SIZE,
        onchain_config=onchain_config,
        offchain_config_version=offchain_config_version,
        offchain_config_len=offchain_config_len,
        offchain_config=offchain_config,
    )

    return (digest)
end

func remove_oracles{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(n : felt):
    if n == 0:
        Aggregator_oracles_len.write(0)
        return ()
    end

    # delete oracle from all maps
    let (signer) = Aggregator_signers_list.read(n)
    Aggregator_signers.write(signer, 0)

    let (transmitter) = Aggregator_transmitters_list.read(n)
    Aggregator_transmitters.write(transmitter, Oracle(index=0, payment_juels=0))

    return remove_oracles(n - 1)
end

func add_oracles{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    oracles : OracleConfig*, index : felt, len : felt, latest_round_id : felt
):
    if len == 0:
        Aggregator_oracles_len.write(index)
        return ()
    end

    # NOTE: index should start with 1 here because storage is 0-initialized.
    # That way signers(pkey) => 0 indicates "not present"
    let index = index + 1

    # Check for duplicates
    let (existing_signer) = Aggregator_signers.read(oracles.signer)
    with_attr error_message("repeated signer"):
        assert existing_signer = 0
    end

    let (existing_transmitter : Oracle) = Aggregator_transmitters.read(oracles.transmitter)
    with_attr error_message("repeated transmitter"):
        assert existing_transmitter.index = 0
    end

    Aggregator_signers.write(oracles.signer, index)
    Aggregator_signers_list.write(index, oracles.signer)

    Aggregator_transmitters.write(oracles.transmitter, Oracle(index=index, payment_juels=0))
    Aggregator_transmitters_list.write(index, oracles.transmitter)

    reward_from_aggregator_round_id_.write(index, latest_round_id)

    return add_oracles(oracles + OracleConfig.SIZE, index, len - 1, latest_round_id)
end

const DIGEST_MASK = 2 ** (252 - 12) - 1
const PREFIX = 4 * 2 ** (252 - 12)

func config_digest_from_data{pedersen_ptr : HashBuiltin*, bitwise_ptr : BitwiseBuiltin*}(
    chain_id : felt,
    contract_address : felt,
    config_count : felt,
    oracles_len : felt,
    oracles : OracleConfig*,
    f : felt,
    onchain_config_len : felt,
    onchain_config : felt*,
    offchain_config_version : felt,
    offchain_config_len : felt,
    offchain_config : felt*,
) -> (hash : felt):
    let hash_ptr = pedersen_ptr
    with hash_ptr:
        let (hash_state_ptr) = hash_init()
        let (hash_state_ptr) = hash_update_single(hash_state_ptr, chain_id)
        let (hash_state_ptr) = hash_update_single(hash_state_ptr, contract_address)
        let (hash_state_ptr) = hash_update_single(hash_state_ptr, config_count)
        let (hash_state_ptr) = hash_update_single(hash_state_ptr, oracles_len)
        let (hash_state_ptr) = hash_update(hash_state_ptr, oracles, oracles_len * OracleConfig.SIZE)
        let (hash_state_ptr) = hash_update_single(hash_state_ptr, f)
        let (hash_state_ptr) = hash_update_single(hash_state_ptr, onchain_config_len)
        let (hash_state_ptr) = hash_update(hash_state_ptr, onchain_config, onchain_config_len)
        let (hash_state_ptr) = hash_update_single(hash_state_ptr, offchain_config_version)
        let (hash_state_ptr) = hash_update_single(hash_state_ptr, offchain_config_len)
        let (hash_state_ptr) = hash_update(hash_state_ptr, offchain_config, offchain_config_len)

        let (hash) = hash_finalize(hash_state_ptr)

        # clamp the first two bytes with the config digest prefix
        let (masked) = bitwise_and(hash, DIGEST_MASK)
        let hash = masked + PREFIX

        let pedersen_ptr = hash_ptr
        return (hash=hash)
    end
end

@view
func latest_config_details{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    config_count : felt, block_number : felt, config_digest : felt
):
    let (config_count) = Aggregator_config_count.read()
    let (block_number) = Aggregator_latest_config_block_number.read()
    let (config_digest) = Aggregator_latest_config_digest.read()
    return (config_count=config_count, block_number=block_number, config_digest=config_digest)
end

@view
func transmitters{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    transmitters_len : felt, transmitters : felt*
):
    alloc_locals

    let (result : felt*) = alloc()
    let (len) = Aggregator_oracles_len.read()

    transmitters_inner(len, 0, result)

    return (transmitters_len=len, transmitters=result)
end

# unroll transmitter list into a continuous array
func transmitters_inner{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    len : felt, index : felt, result : felt*
):
    if len == 0:
        return ()
    end

    let index = index + 1

    let (transmitter) = Aggregator_transmitters_list.read(index)
    assert result[0] = transmitter

    return transmitters_inner(len - 1, index, result + 1)
end

# --- Transmission ---

struct Signature:
    member r : felt
    member s : felt
    member public_key : felt
end

struct ReportContext:
    member config_digest : felt
    member epoch_and_round : felt
    member extra_hash : felt
end

@external
func transmit{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    ecdsa_ptr : SignatureBuiltin*,
    bitwise_ptr : BitwiseBuiltin*,
    range_check_ptr,
}(
    report_context : ReportContext,
    observation_timestamp : felt,
    observers : felt,
    observations_len : felt,
    observations : felt*,
    juels_per_fee_coin : felt,
    gas_price : felt,
    signatures_len : felt,
    signatures : Signature*,
):
    alloc_locals

    let (epoch_and_round) = Aggregator_latest_epoch_and_round.read()
    with_attr error_message("stale report"):
        assert_lt(epoch_and_round, report_context.epoch_and_round)
    end

    # validate transmitter
    let (caller) = get_caller_address()
    let (oracle : Oracle) = Aggregator_transmitters.read(caller)
    assert_not_zero(oracle.index)  # 0 index = uninitialized

    # Validate config digest matches latest_config_digest
    let (config_digest) = Aggregator_latest_config_digest.read()
    with_attr error_message("config digest mismatch"):
        assert report_context.config_digest = config_digest
    end

    let (f) = Aggregator_f.read()
    with_attr error_message("wrong number of signatures f={f}"):
        assert signatures_len = (f + 1)
    end

    let (msg) = hash_report(
        report_context,
        observation_timestamp,
        observers,
        observations_len,
        observations,
        juels_per_fee_coin,
        gas_price,
    )
    verify_signatures(msg, signatures, signatures_len, signed_count=0)

    # report():

    assert_nn_le(observations_len, MAX_ORACLES)  # len <= MAX_ORACLES
    assert_lt(f, observations_len)  # f < len

    Aggregator_latest_epoch_and_round.write(report_context.epoch_and_round)

    let (median_idx : felt, _) = unsigned_div_rem(observations_len, 2)
    let median = observations[median_idx]

    # Check abs(median) is in i128 range.
    # NOTE: (assert_lt_felt(-i128::MAX, median) doesn't work correctly so we have to use abs!)
    let (value) = abs_value(median)
    with_attr error_message("value not in int128 range: {median}"):
        assert_lt_felt(value, INT128_MAX)
    end

    # Validate median in min-max range
    let (answer_range : Range) = Aggregator_answer_range.read()
    assert_in_range(median, answer_range.min, answer_range.max)

    let (local prev_round_id) = Aggregator_latest_aggregator_round_id.read()
    # let (prev_round_id) = Aggregator_latest_aggregator_round_id.read()
    let round_id = prev_round_id + 1
    Aggregator_latest_aggregator_round_id.write(round_id)

    let (timestamp : felt) = get_block_timestamp()
    let (block_num : felt) = get_block_number()

    # write to storage
    Aggregator_transmissions.write(
        round_id,
        Transmission(
        answer=median,
        block_num=block_num,
        observation_timestamp=observation_timestamp,
        transmission_timestamp=timestamp,
        ),
    )

    # NOTE: Usually validating via validator would happen here, currently disabled

    let (billing : Billing) = Aggregator_billing.read()

    let (reimbursement_juels) = calculate_reimbursement(
        juels_per_fee_coin, signatures_len, gas_price, billing
    )

    # end report()

    NewTransmission.emit(
        round_id=round_id,
        answer=median,
        transmitter=caller,
        observation_timestamp=observation_timestamp,
        observers=observers,
        observations_len=observations_len,
        observations=observations,
        juels_per_fee_coin=juels_per_fee_coin,
        gas_price=gas_price,
        config_digest=report_context.config_digest,
        epoch_and_round=report_context.epoch_and_round,
        reimbursement=reimbursement_juels,
    )

    # pay transmitter
    let payment = reimbursement_juels + (billing.transmission_payment_gjuels * GIGA)
    # TODO: check overflow

    Aggregator_transmitters.write(
        caller,
        Oracle(
        index=oracle.index,
        payment_juels=oracle.payment_juels + payment
        ),
    )

    return ()
end

func hash_report{pedersen_ptr : HashBuiltin*}(
    report_context : ReportContext,
    observation_timestamp : felt,
    observers : felt,
    observations_len : felt,
    observations : felt*,
    juels_per_fee_coin : felt,
    gas_price : felt,
) -> (hash : felt):
    let hash_ptr = pedersen_ptr
    with hash_ptr:
        let (hash_state_ptr) = hash_init()
        let (hash_state_ptr) = hash_update_single(hash_state_ptr, report_context.config_digest)
        let (hash_state_ptr) = hash_update_single(hash_state_ptr, report_context.epoch_and_round)
        let (hash_state_ptr) = hash_update_single(hash_state_ptr, report_context.extra_hash)
        let (hash_state_ptr) = hash_update_single(hash_state_ptr, observation_timestamp)
        let (hash_state_ptr) = hash_update_single(hash_state_ptr, observers)
        let (hash_state_ptr) = hash_update_single(hash_state_ptr, observations_len)
        let (hash_state_ptr) = hash_update(hash_state_ptr, observations, observations_len)
        let (hash_state_ptr) = hash_update_single(hash_state_ptr, juels_per_fee_coin)
        let (hash_state_ptr) = hash_update_single(hash_state_ptr, gas_price)

        let (hash) = hash_finalize(hash_state_ptr)
        let pedersen_ptr = hash_ptr
        return (hash=hash)
    end
end

func verify_signatures{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    ecdsa_ptr : SignatureBuiltin*,
    bitwise_ptr : BitwiseBuiltin*,
    range_check_ptr,
}(msg : felt, signatures : Signature*, signatures_len : felt, signed_count : felt):
    alloc_locals

    # 'signed_count' is used for tracking duplicate signatures
    if signatures_len == 0:
        # Check all signatures are unique (we only saw each pubkey once)
        let (masked) = bitwise_and(
            signed_count, 0x01010101010101010101010101010101010101010101010101010101010101
        )
        with_attr error_message("duplicate signer"):
            assert signed_count = masked
        end
        return ()
    end

    let signature = signatures[0]

    # Validate the signer key actually belongs to an oracle
    let (index) = Aggregator_signers.read(signature.public_key)
    with_attr error_message("invalid signer {signature.public_key}"):
        assert_not_zero(index)  # 0 index = uninitialized
    end

    verify_ecdsa_signature(
        message=msg,
        public_key=signature.public_key,
        signature_r=signature.r,
        signature_s=signature.s,
    )

    # TODO: Using shifts here might be expensive due to pow()?
    # evaluate using alloc() to allocate a signed_count[oracles_len] instead

    # signed_count + 1 << (8 * index)
    let (shift) = pow(2, 8 * index)
    let signed_count = signed_count + shift

    return verify_signatures(msg, signatures + Signature.SIZE, signatures_len - 1, signed_count)
end

@view
func latest_transmission_details{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    ) -> (
    config_digest : felt, epoch_and_round : felt, latest_answer : felt, latest_timestamp : felt
):
    let (config_digest) = Aggregator_latest_config_digest.read()
    let (latest_round_id) = Aggregator_latest_aggregator_round_id.read()
    let (epoch_and_round) = Aggregator_latest_epoch_and_round.read()
    let (transmission : Transmission) = Aggregator_transmissions.read(latest_round_id)

    return (
        config_digest=config_digest,
        epoch_and_round=epoch_and_round,
        latest_answer=transmission.answer,
        latest_timestamp=transmission.transmission_timestamp,
    )
end

# --- RequestNewRound

# --- Queries

# Read access helper
func require_access{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}():
    let (address) = get_caller_address()
    SimpleReadAccessController.check_access(address)

    return ()
end

@view
func description{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    description : felt
):
    require_access()
    let (description) = Aggregator_description.read()
    return (description)
end

@view
func decimals{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    decimals : felt
):
    require_access()
    let (decimals) = Aggregator_decimals.read()
    return (decimals)
end

@view
func round_data{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    round_id : felt
) -> (round : Round):
    require_access()
    # TODO: assert round_id fits in u32

    let (transmission : Transmission) = Aggregator_transmissions.read(round_id)

    let round = Round(
        round_id=round_id,
        answer=transmission.answer,
        block_num=transmission.block_num,
        started_at=transmission.observation_timestamp,
        updated_at=transmission.transmission_timestamp,
    )
    return (round)
end

@view
func latest_round_data{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    round : Round
):
    let (latest_round_id) = Aggregator_latest_aggregator_round_id.read()
    let (transmission : Transmission) = Aggregator_transmissions.read(latest_round_id)

    let round = Round(
        round_id=latest_round_id,
        answer=transmission.answer,
        block_num=transmission.block_num,
        started_at=transmission.observation_timestamp,
        updated_at=transmission.transmission_timestamp,
    )
    return (round)
end

# --- Set LINK Token

@storage_var
func Aggregator_link_token() -> (token : felt):
end

@event
func LinkTokenSet(old_link_token : felt, new_link_token : felt):
end

@external
func set_link_token{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    link_token : felt, recipient : felt
):
    alloc_locals
    Ownable.assert_only_owner()

    let (old_token) = Aggregator_link_token.read()
    if link_token == old_token:
        return ()
    end

    let (contract_address) = get_contract_address()

    # call balanceOf as a sanity check to confirm we're talking to a token
    IERC20.balanceOf(contract_address=link_token, account=contract_address)

    pay_oracles()

    # transfer remaining balance to recipient
    let (amount : Uint256) = IERC20.balanceOf(contract_address=link_token, account=contract_address)
    IERC20.transfer(contract_address=old_token, recipient=recipient, amount=amount)

    Aggregator_link_token.write(link_token)

    LinkTokenSet.emit(old_link_token=old_token, new_link_token=link_token)

    return ()
end

@view
func link_token{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    link_token : felt
):
    let (link_token) = Aggregator_link_token.read()
    return (link_token)
end

# --- Billing Access Controller

@storage_var
func Aggregator_billing_access_controller() -> (access_controller : felt):
end

@event
func BillingAccessControllerSet(old_controller : felt, new_controller : felt):
end

@external
func set_billing_access_controller{
    syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr
}(access_controller : felt):
    Ownable.assert_only_owner()

    let (old_controller) = Aggregator_billing_access_controller.read()
    if access_controller != old_controller:
        Aggregator_billing_access_controller.write(access_controller)

        BillingAccessControllerSet.emit(
            old_controller=old_controller, new_controller=access_controller
        )

        return ()
    end

    return ()
end

@view
func billing_access_controller{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    ) -> (access_controller : felt):
    let (access_controller) = Aggregator_billing_access_controller.read()
    return (access_controller)
end

# --- Billing Config

struct Billing:
    # TODO: use a single felt via (observation_payment, transmission_payment) = split_felt()?
    member observation_payment_gjuels : felt
    member transmission_payment_gjuels : felt
    member gas_base : felt
    member gas_per_signature : felt
end

@storage_var
func Aggregator_billing() -> (config : Billing):
end

@view
func billing{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    config : Billing
):
    let (config : Billing) = Aggregator_billing.read()
    return (config)
end

@event
func BillingSet(config : Billing):
end

@external
func set_billing{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    config : Billing
):
    has_billing_access()

    # Pay out oracles using existing settings for rounds up to now
    pay_oracles()

    # check payment value ranges within u32 bounds
    assert_nn_le(config.observation_payment_gjuels, UINT32_MAX)
    assert_nn_le(config.transmission_payment_gjuels, UINT32_MAX)
    assert_nn_le(config.gas_base, UINT32_MAX)
    assert_nn_le(config.gas_per_signature, UINT32_MAX)

    Aggregator_billing.write(config)

    BillingSet.emit(config=config)

    return ()
end

func has_billing_access{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}():
    let (caller) = get_caller_address()
    let (owner) = Ownable.get_owner()

    # owner always has access
    if caller == owner:
        return ()
    end

    let (access_controller) = Aggregator_billing_access_controller.read()

    IAccessController.check_access(contract_address=access_controller, user=caller)
    return ()
end

# --- Payments and Withdrawals

@event
func OraclePaid(transmitter : felt, payee : felt, amount : Uint256, link_token : felt):
end

@external
func withdraw_payment{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    transmitter : felt
):
    alloc_locals
    let (caller) = get_caller_address()
    let (payee) = Aggregator_payees.read(transmitter)
    with_attr error_message("only payee can withdraw"):
        assert caller = payee
    end

    let (latest_round_id) = Aggregator_latest_aggregator_round_id.read()
    let (link_token) = Aggregator_link_token.read()
    pay_oracle(transmitter, latest_round_id, link_token)
    return ()
end

func _owed_payment{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    oracle : Oracle
) -> (amount : felt):
    if oracle.index == 0:
        return (0)
    end

    let (billing : Billing) = Aggregator_billing.read()

    let (latest_round_id) = Aggregator_latest_aggregator_round_id.read()
    let (from_round_id) = reward_from_aggregator_round_id_.read(oracle.index)
    let rounds = latest_round_id - from_round_id

    let amount = (rounds * billing.observation_payment_gjuels * GIGA) + oracle.payment_juels
    return (amount)
end

@external
func owed_payment{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    transmitter : felt
) -> (amount : felt):
    let (oracle : Oracle) = Aggregator_transmitters.read(transmitter)
    let (amount : felt) = _owed_payment(oracle)
    return (amount)
end

func pay_oracle{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    transmitter : felt, latest_round_id : felt, link_token : felt
):
    alloc_locals

    let (oracle : Oracle) = Aggregator_transmitters.read(transmitter)

    if oracle.index == 0:
        return ()
    end

    let (amount_ : felt) = _owed_payment(oracle)
    assert_nn(amount_)

    # if zero, fastpath return to avoid empty transfers
    if amount_ == 0:
        return ()
    end

    let (amount : Uint256) = felt_to_uint256(amount_)
    let (payee) = Aggregator_payees.read(transmitter)

    IERC20.transfer(contract_address=link_token, recipient=payee, amount=amount)

    # Reset payment
    reward_from_aggregator_round_id_.write(oracle.index, latest_round_id)
    Aggregator_transmitters.write(transmitter, Oracle(index=oracle.index, payment_juels=0))

    OraclePaid.emit(transmitter=transmitter, payee=payee, amount=amount, link_token=link_token)

    return ()
end

func pay_oracles{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}():
    let (len) = Aggregator_oracles_len.read()
    let (latest_round_id) = Aggregator_latest_aggregator_round_id.read()
    let (link_token) = Aggregator_link_token.read()
    pay_oracles_(len, latest_round_id, link_token)
    return ()
end

func pay_oracles_{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    index : felt, latest_round_id : felt, link_token : felt
):
    if index == 0:
        return ()
    end

    let (transmitter) = Aggregator_transmitters_list.read(index)
    pay_oracle(transmitter, latest_round_id, link_token)

    return pay_oracles_(index - 1, latest_round_id, link_token)
end

@external
func withdraw_funds{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    recipient : felt, amount : Uint256
):
    alloc_locals
    has_billing_access()

    let (link_token) = Aggregator_link_token.read()
    let (contract_address) = get_contract_address()

    let (link_due) = total_link_due()
    let (balance : Uint256) = IERC20.balanceOf(
        contract_address=link_token, account=contract_address
    )

    let (link_due_uint256 : Uint256) = felt_to_uint256(link_due)
    let (res) = uint256_le(link_due_uint256, balance)
    with_attr error_message("Total amount due exceeds the balance"):
        assert res = 1
    end

    let (available : Uint256) = uint256_sub(balance, link_due_uint256)

    let (less_available : felt) = uint256_lt(available, amount)
    if less_available == TRUE:
        # Transfer as much as there is available
        IERC20.transfer(contract_address=link_token, recipient=recipient, amount=available)
    else:
        IERC20.transfer(contract_address=link_token, recipient=recipient, amount=amount)
    end

    return ()
end

func total_link_due{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    due : felt
):
    let (len) = Aggregator_oracles_len.read()
    let (latest_round_id) = Aggregator_latest_aggregator_round_id.read()

    let (amount) = total_link_due_(len, latest_round_id, 0, 0)
    return (amount)
end

func total_link_due_{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    index : felt, latest_round_id : felt, total_rounds : felt, payments_juels : felt
) -> (due : felt):
    if index == 0:
        let (billing : Billing) = Aggregator_billing.read()
        let amount = (total_rounds * billing.observation_payment_gjuels * GIGA) + payments_juels
        return (amount)
    end

    let (transmitter) = Aggregator_transmitters_list.read(index)
    let (oracle : Oracle) = Aggregator_transmitters.read(transmitter)
    assert_not_zero(oracle.index)  # 0 == undefined

    let (from_round_id) = reward_from_aggregator_round_id_.read(oracle.index)
    let rounds = latest_round_id - from_round_id

    let total_rounds = total_rounds + rounds
    let payments_juels = payments_juels + oracle.payment_juels

    return total_link_due_(index - 1, latest_round_id, total_rounds, payments_juels)
end

@view
func link_available_for_payment{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    ) -> (available : felt):
    alloc_locals
    let (link_token) = Aggregator_link_token.read()
    let (contract_address) = get_contract_address()

    let (balance_ : Uint256) = IERC20.balanceOf(
        contract_address=link_token, account=contract_address
    )
    # entire link supply fits into u96 so this should not fail
    let (balance) = uint256_to_felt(balance_)

    let (due) = total_link_due()
    with_attr error_message("Total amount due exceeds the balance"):
        assert_le(balance, due)
    end
    let amount = balance - due

    return (available=amount)
end

# --- Transmitter Payment

const MARGIN = 115

func calculate_reimbursement{range_check_ptr}(
    juels_per_fee_coin : felt, signature_count : felt, gas_price : felt, config : Billing
) -> (amount_juels : felt):
    # Based on estimateFee (f=1 14977, f=2 14989, f=3 15002 f=4 15014 f=5 15027, count = f+1)
    # NOTE: seems a bit odd since each ecdsa is supposed to be 25.6 gas: https://docs.starknet.io/docs/Fees/fee-mechanism/
    # gas_base = 14951, gas_per_signature = 13
    let exact_gas = config.gas_base + (signature_count * config.gas_per_signature)
    let (gas : felt, _) = unsigned_div_rem(exact_gas * MARGIN, 100)  # scale to 115% for some margin
    let amount = gas * gas_price
    let amount_juels = amount * juels_per_fee_coin
    return (amount_juels)
end

# --- Payee Management

@storage_var
func Aggregator_payees(transmitter : felt) -> (payment_address : felt):
end

@storage_var
func Aggregator_proposed_payees(transmitter : felt) -> (payment_address : felt):
end

@event
func PayeeshipTransferRequested(transmitter : felt, current : felt, proposed : felt):
end

@event
func PayeeshipTransferred(transmitter : felt, previous : felt, current : felt):
end

struct PayeeConfig:
    member transmitter : felt
    member payee : felt
end

@external
func set_payees{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    payees_len : felt, payees : PayeeConfig*
):
    Ownable.assert_only_owner()

    set_payee(payees, payees_len)

    return ()
end

# Returns 1 if value == 0. Returns 1 otherwise.
func is_zero(value) -> (res):
    if value == 0:
        return (res=1)
    end

    return (res=0)
end

func set_payee{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    payees : PayeeConfig*, len : felt
):
    if len == 0:
        return ()
    end

    let (current_payee) = Aggregator_payees.read(payees.transmitter)

    # a more convoluted way of saying
    # require(current_payee == 0 || current_payee == payee, "payee already set")
    let (is_unset) = is_zero(current_payee)
    let (is_same) = is_zero(current_payee - payees.payee)
    with_attr error_message("payee already set"):
        assert (is_unset - 1) * (is_same - 1) = 0
    end

    Aggregator_payees.write(payees.transmitter, payees.payee)

    PayeeshipTransferred.emit(
        transmitter=payees.transmitter, previous=current_payee, current=payees.payee
    )

    return set_payee(payees + PayeeConfig.SIZE, len - 1)
end

@external
func transfer_payeeship{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    transmitter : felt, proposed : felt
):
    with_attr error_message("cannot transfer payeeship to zero address"):
        assert_not_zero(proposed)
    end
    let (caller) = get_caller_address()
    let (payee) = Aggregator_payees.read(transmitter)
    with_attr error_message("only current payee can update"):
        assert caller = payee
    end
    with_attr error_message("cannot transfer to self"):
        assert_not_equal(caller, proposed)
    end

    Aggregator_proposed_payees.write(transmitter, proposed)

    PayeeshipTransferRequested.emit(transmitter=transmitter, current=payee, proposed=proposed)

    return ()
end

@external
func accept_payeeship{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    transmitter : felt
):
    let (proposed) = Aggregator_proposed_payees.read(transmitter)
    let (caller) = get_caller_address()
    # caller cannot be zero address to avoid overwriting owner when proposed_owner is not set
    with_attr error_message("caller is the zero address"):
        assert_not_zero(caller)
    end
    with_attr error_message("only proposed payee can accept"):
        assert caller = proposed
    end

    let (previous) = Aggregator_payees.read(transmitter)
    Aggregator_payees.write(transmitter, caller)
    Aggregator_proposed_payees.write(transmitter, 0)

    PayeeshipTransferred.emit(transmitter=transmitter, previous=previous, current=caller)

    return ()
end

@view
func type_and_version() -> (meta : felt):
    return ('ocr2/aggregator.cairo 1.0.0')
end
