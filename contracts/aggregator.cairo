%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin, SignatureBuiltin, BitwiseBuiltin
from starkware.cairo.common.alloc import alloc
from starkware.cairo.common.hash import hash2
from starkware.cairo.common.hash_state import (
    hash_init, hash_finalize, hash_update, hash_update_single
)
from starkware.cairo.common.signature import verify_ecdsa_signature
from starkware.cairo.common.bitwise import bitwise_and
from starkware.cairo.common.math import (
    split_felt,
    assert_not_zero, assert_not_equal, assert_lt, assert_nn_le, assert_nn, assert_in_range, unsigned_div_rem
)
from starkware.cairo.common.math_cmp import (
    is_not_zero,
)
from starkware.cairo.common.pow import pow
from starkware.cairo.common.uint256 import (
    Uint256,
)
from starkware.cairo.common.bool import TRUE, FALSE

from starkware.starknet.common.syscalls import (
    get_caller_address,
    get_contract_address,
    get_block_timestamp,
    get_block_number
)

from openzeppelin.utils.constants import UINT8_MAX

from openzeppelin.token.erc20.interfaces.IERC20 import IERC20

from contracts.ownable import (
    Ownable_initializer,
    Ownable_only_owner,
    Ownable_get_owner,
    Ownable_transfer_ownership,
    Ownable_accept_ownership
)

# ---

const MAX_ORACLES = 31

func felt_to_uint256{range_check_ptr}(x) -> (uint_x : Uint256):
    let (high, low) = split_felt(x)
    return (Uint256(low=low, high=high))
end

@storage_var
func link_token_() -> (token: felt):
end

# Maximum number of faulty oracles
@storage_var
func f_() -> (f: felt):
end

@storage_var
func latest_epoch_and_round_() -> (res: felt):
end

@storage_var
func latest_aggregator_round_id_() -> (round_id: felt):
end

@storage_var
func answer_range_() -> (range: (felt, felt)):
end

@storage_var
func decimals_() -> (decimals: felt):
end

@storage_var
func description_() -> (description: felt):
end

#

@storage_var
func latest_config_block_number_() -> (block: felt):
end

@storage_var
func config_count_() -> (count: felt):
end

@storage_var
func latest_config_digest_() -> (digest: felt):
end

# struct Oracle:
#   # TODO: payment amount, from_round_id
# end

@storage_var
func oracles_len_() -> (len: felt):
end

# @storage_var
# func oracles_(index: felt) -> (oracle: Oracle):
# end

# TODO: also store payment here?
@storage_var
func transmitters_(pkey: felt) -> (index: felt):
end

@storage_var
func signers_(pkey: felt) -> (index: felt):
end

@storage_var
func signers_list_(index: felt) -> (pkey: felt):
end

@storage_var
func transmitters_list_(index: felt) -> (pkey: felt):
end

@storage_var
func reward_from_aggregator_round_id_(index: felt) -> (round_id: felt):
end

# ---

struct Transmission:
    member answer: felt
    member block_num: felt
    member observation_timestamp: felt
    member transmission_timestamp: felt
end

@storage_var
func transmissions_(round_id: felt) -> (transmission: Transmission):
end

# ---

@constructor
func constructor{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}(
    owner: felt,
    link: felt,
    min_answer: felt,
    max_answer: felt,
    billing_access_controller: felt,
    decimals: felt,
    description: felt
):
    Ownable_initializer(owner)
    link_token_.write(link)
    billing_access_controller_.write(billing_access_controller)

    assert_lt(min_answer, max_answer)
    answer_range_.write((min_answer, max_answer))

    with_attr error_message("decimals exceed 2^8"):
        assert_lt(decimals, UINT8_MAX)
    end
    decimals_.write(decimals)
    description_.write(description)
    # TODO: initialize vars to defaults
    return ()
end

# --- Ownership ---

@view
func owner{pedersen_ptr : HashBuiltin*, syscall_ptr : felt*, range_check_ptr}() -> (owner : felt):
    let (owner) = Ownable_get_owner()
    return (owner=owner)
end

@external
func transfer_ownership{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr
}(new_owner: felt) -> ():
    return Ownable_transfer_ownership(new_owner)
end

@external
func accept_ownership{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr
}() -> (new_owner: felt):
    return Ownable_accept_ownership()
end

# --- Validation ---

@contract_interface
namespace IValidator:
    func validate(prev_round_id: felt, prev_answer: felt, round_id: felt, answer: felt):
    end
end

# TODO: can't set gas limit
@storage_var
func validator_() -> (validator: felt):
end

@view
func validator_config{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}() -> (validator: felt):
    let (validator) = validator_.read()
    return (validator)
end

@external
func set_validator_config{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}(validator: felt):
    Ownable_only_owner()
    # TODO: use openzeppelin's ERC165 to validate
    validator_.write(validator)

    # TODO: emit event

    return ()
end

# --- Configuration

@event
func config_set(
    previous_config_block_number: felt,
    latest_config_digest: felt,
    config_count: felt,
    oracles_len: felt,
    oracles: OracleConfig*,
    f: felt,
    onchain_config: felt, # TODO
    offchain_config_version: felt,
    offchain_config_len: felt,
    offchain_config: felt*,
):
end


struct OracleConfig:
    member signer: felt
    member transmitter: felt
end

@external
func set_config{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr
}(
    oracles_len: felt,
    oracles: OracleConfig*,
    f: felt,
    onchain_config: felt, # TODO
    offchain_config_version: felt,
    offchain_config_len: felt,
    offchain_config: felt*,
) -> (digest: felt):
    alloc_locals
    # Ownable_only_owner() TODO: reenable

    assert_nn_le(oracles_len, MAX_ORACLES) # oracles_len <= MAX_ORACLES
    assert_lt(3 * f, oracles_len) # 3 * f < oracles_len
    assert_nn(f) # f is positive

    # pay out existing oracles
    pay_oracles()

    # remove old signers/transmitters
    let (len) = oracles_len_.read()
    remove_oracles(len)

    let (latest_round_id) = latest_aggregator_round_id_.read()

    # add new oracles (also sets oracle_len_)
    add_oracles(oracles, 0, oracles_len, latest_round_id)

    f_.write(f)
    let (block_num : felt) = get_block_number()
    let (prev_block_num) = latest_config_block_number_.read()
    latest_config_block_number_.write(block_num)
    # update config count
    let (config_count) = config_count_.read()
    let config_count = config_count + 1
    config_count_.write(config_count)
    # calculate and store config digest
    let (contract_address) = get_contract_address()
    let (digest) = config_digest_from_data(
        contract_address,
        config_count,
        oracles_len,
        oracles,
        f,
        onchain_config,
        offchain_config_version,
        offchain_config_len,
        offchain_config
    )
    latest_config_digest_.write(digest)

    # reset epoch & round
    latest_epoch_and_round_.write(0)

    config_set.emit(
        previous_config_block_number=prev_block_num,
        latest_config_digest=digest,
        config_count=config_count,
        oracles_len=oracles_len,
        oracles=oracles,
        f=f,
        onchain_config=onchain_config,
        offchain_config_version=offchain_config_version,
        offchain_config_len=offchain_config_len,
        offchain_config=offchain_config,
    )

    return (digest)
end

func remove_oracles{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}(n: felt):
    alloc_locals

    if n == 0:
        oracles_len_.write(0)
        return ()
    end

    # delete oracle from all maps
    let (signer) = signers_list_.read(n)
    signers_.write(signer, 0)

    let (transmitter) = transmitters_list_.read(n)
    transmitters_.write(transmitter, 0)

    return remove_oracles(n - 1)
end

# NOTE: index should start with 1 here because storage is 0-initialized.
# That way signers(pkey) => 0 indicates "not present"
func add_oracles{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}(oracles: OracleConfig*, index: felt, len: felt, latest_round_id: felt):
    alloc_locals

    if len == 0:
        oracles_len_.write(index)
        return ()
    end

    let index = index + 1

    signers_.write(oracles.signer, index)
    signers_list_.write(index, oracles.signer)

    transmitters_.write(oracles.transmitter, index)
    transmitters_list_.write(index, oracles.transmitter)

    reward_from_aggregator_round_id_.write(index, latest_round_id)

    return add_oracles(oracles + OracleConfig.SIZE, index, len - 1, latest_round_id)
end

func config_digest_from_data{
    pedersen_ptr : HashBuiltin*,
}(
    contract_address: felt,
    config_count: felt,
    oracles_len: felt,
    oracles: OracleConfig*,
    f: felt,
    onchain_config: felt, # TODO
    offchain_config_version: felt,
    offchain_config_len: felt,
    offchain_config: felt*,
) -> (hash: felt):
    let hash_ptr = pedersen_ptr
    with hash_ptr:
        let (hash_state_ptr) = hash_init()
        let (hash_state_ptr) = hash_update_single(hash_state_ptr, contract_address)
        let (hash_state_ptr) = hash_update_single(hash_state_ptr, config_count)
        let (hash_state_ptr) = hash_update_single(hash_state_ptr, oracles_len)
        let (hash_state_ptr) = hash_update(hash_state_ptr, oracles, oracles_len * OracleConfig.SIZE)
        let (hash_state_ptr) = hash_update_single(hash_state_ptr, f)
        let (hash_state_ptr) = hash_update_single(hash_state_ptr, onchain_config)
        let (hash_state_ptr) = hash_update_single(hash_state_ptr, offchain_config_version)
        let (hash_state_ptr) = hash_update_single(hash_state_ptr, offchain_config_len)
        let (hash_state_ptr) = hash_update(hash_state_ptr, offchain_config, offchain_config_len)

        let (hash) = hash_finalize(hash_state_ptr)
        let pedersen_ptr = hash_ptr
        return (hash=hash)
    end
end

@view
func latest_config_details{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}() -> (
    config_count: felt,
    block_number: felt,
    config_digest: felt
):
    let (config_count) = config_count_.read()
    let (block_number) = latest_config_block_number_.read()
    let (config_digest) = latest_config_digest_.read()
    return(config_count=config_count, block_number=block_number, config_digest=config_digest)
end

@view
func transmitters{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}() -> (transmitters_len: felt, transmitters: felt*):
    alloc_locals

    let (result: felt*) = alloc()
    let (len) = oracles_len_.read()

    transmitters_inner(len, 0, result)

    return (transmitters_len=len, transmitters=result)
end

# unroll transmitter list into a continuous array
func transmitters_inner{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}(len: felt, index: felt, result: felt*):
    if len == 0:
        return ()
    end

    let index = index + 1

    let (transmitter) = transmitters_list_.read(index)
    assert result[0] = transmitter

    return transmitters_inner(len, index, result + 1)
end

# --- Transmission ---

@event
func new_transmission(
    round_id: felt,
    answer: felt,
    transmitter: felt,
    observations_timestamp: felt,
    observers: felt,
    observations_len: felt,
    observations: felt*,
    juels_per_fee_coin: felt,
    config_digest: felt,
    epoch_and_round: felt, # TODO: split?
    reimbursement: Uint256,
):
end

struct Signature:
    member r : felt
    member s : felt
    # TODO: can further compress by using signer index instead of pubkey?
    # TODO: observers[i] = n => signers[n] => public_key
    member public_key: felt
end

struct ReportContext:
    member config_digest : felt
    member epoch_and_round : felt
    member extra_hash : felt
end

# TODO we can base64 encode inputs, but we could also pre-split the inputs (so instead of a binary report,
# it's already split into observers, len and observations). Encoding would shrink the input size since each observation
# wouldn't have to be felt-sized.
@external
func transmit{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    ecdsa_ptr : SignatureBuiltin*,
    bitwise_ptr : BitwiseBuiltin*,
    range_check_ptr
}(
    # TODO: timestamp & juels_per_fee_coin
    report_context: ReportContext,
    observers: felt,
    observations_len: felt,
    observations: felt*,
    signatures_len: felt,
    signatures: Signature*,
):
    alloc_locals

    let (epoch_and_round) = latest_epoch_and_round_.read()
    with_attr error_message("stale report"):
        assert_lt(epoch_and_round, report_context.epoch_and_round)
    end

    # validate transmitter
    let (caller) = get_caller_address()
    let (oracle_idx) = transmitters_.read(caller)
    assert_not_equal(oracle_idx, 0) # 0 index = uninitialized
    # ERROR: caller seems to be the account contract address, not the underlying transmitter key

    # Validate config digest matches latest_config_digest
    let (config_digest) = latest_config_digest_.read()
    with_attr error_message("config digest mismatch"):
        assert report_context.config_digest = config_digest
    end

    let (f) = f_.read()
    with_attr error_message("wrong number of signatures f={f}"):
        assert signatures_len = (f + 1)
    end

    let (msg) = hash_report(report_context, observers, observations_len, observations)
    verify_signatures(msg, signatures, signatures_len, signed_count=0)

    # report():

    assert_nn_le(observations_len, MAX_ORACLES) # len <= MAX_ORACLES
    assert_lt(f, observations_len) # f < len

    latest_epoch_and_round_.write(report_context.epoch_and_round)

    let (median_idx : felt, _) = unsigned_div_rem(observations_len, 2)
    let median = observations[median_idx]

    # TODO: assert inside i192 range

    # Validate median in min-max range
    let (answer_range) = answer_range_.read()
    assert_in_range(median, answer_range[0], answer_range[1])

    let (prev_round_id) = latest_aggregator_round_id_.read()
    let round_id = prev_round_id + 1
    latest_aggregator_round_id_.write(round_id)

    let (timestamp : felt) = get_block_timestamp()
    let (block_num : felt) = get_block_number()

    # write to storage
    transmissions_.write(round_id, Transmission(
        answer=median,
        block_num=block_num,
        observation_timestamp=1, # TODO:
        transmission_timestamp=timestamp,
    ))

    # validate via validator
    let (validator) = validator_.read()

    if validator != 0:
        let (prev_transmission) = transmissions_.read(prev_round_id)
        IValidator.validate(
            contract_address=validator,
            prev_round_id=prev_round_id,
            prev_answer=prev_transmission.answer,
            round_id=round_id,
            answer=median
        )

        tempvar syscall_ptr = syscall_ptr
        tempvar range_check_ptr = range_check_ptr
        tempvar pedersen_ptr = pedersen_ptr
    else:
        tempvar syscall_ptr = syscall_ptr
        tempvar range_check_ptr = range_check_ptr
        tempvar pedersen_ptr = pedersen_ptr
    end
    tempvar syscall_ptr = syscall_ptr
    tempvar range_check_ptr = range_check_ptr
    tempvar pedersen_ptr = pedersen_ptr # TODO: ??? compilation seems to fail without this follow-up

    # TODO: calculate reimbursement
    let (reimbursement) = calculate_reimbursement()

    # end report()

    new_transmission.emit(
        round_id=round_id,
        answer=median,
        transmitter=caller,
        observations_timestamp=1, # TODO:
        observers=observers,
        observations_len=observations_len,
        observations=observations,
        juels_per_fee_coin=1, # TODO
        config_digest=report_context.config_digest,
        epoch_and_round=report_context.epoch_and_round,
        reimbursement=reimbursement,
    )

    # pay transmitter

    return ()
end

func hash_report{
    pedersen_ptr : HashBuiltin*,
}(
    report_context: ReportContext,
    observers: felt,
    observations_len: felt,
    observations: felt*,
) -> (hash: felt):
    let hash_ptr = pedersen_ptr
    with hash_ptr:
        let (hash_state_ptr) = hash_init()
        # TODO: does hash_update(hash_state_ptr, cast(report_context, felt), ReportContext.SIZE) work?
        let (hash_state_ptr) = hash_update_single(hash_state_ptr, report_context.config_digest)
        let (hash_state_ptr) = hash_update_single(hash_state_ptr, report_context.epoch_and_round)
        let (hash_state_ptr) = hash_update_single(hash_state_ptr, report_context.extra_hash)
        let (hash_state_ptr) = hash_update_single(hash_state_ptr, observers)
        let (hash_state_ptr) = hash_update_single(hash_state_ptr, observations_len)
        let (hash_state_ptr) = hash_update(hash_state_ptr, observations, observations_len)

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
}(
    msg: felt,
    signatures: Signature*,
    signatures_len: felt,
    signed_count: felt # used for tracking duplicate signatures
):
    alloc_locals

    if signatures_len == 0:
        # Check all signatures are unique (we only saw each pubkey once)
        let (masked) = bitwise_and(
            signed_count,
            0x01010101010101010101010101010101010101010101010101010101010101
        )
        with_attr error_message("duplicate signer"):
            assert signed_count = masked
        end
        return ()
    end

    let signature = signatures[0]

    # Validate the signer key actually belongs to an oracle
    let (index) = signers_.read(signature.public_key)
    with_attr error_message("invalid signer {signature.public_key}"):
        assert_not_equal(index, 0) # 0 index = uninitialized
    end

    verify_ecdsa_signature(
        message=msg,
        public_key=signature.public_key,
        signature_r=signature.r,
        signature_s=signature.s
    )

    # TODO: Using shifts here might be expensive due to pow()?

    # signed_count + 1 << (8 * index)
    let (shift) = pow(2, 8 * index)
    let signed_count = signed_count + shift

    return verify_signatures(
        msg,
        signatures + Signature.SIZE,
        signatures_len - 1,
        signed_count
    )
end

# --- RequestNewRound

# --- Queries

@view
func description{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}() -> (description: felt):
    let (description) = description_.read()
    return (description)
end

@view
func decimals{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}() -> (decimals: felt):
    let (decimals) = decimals_.read()
    return (decimals)
end

struct Round:
    member round_id: felt
    member answer: felt
    member block_num: felt
    member observation_timestamp: felt
    member transmission_timestamp: felt
end

@view
func round_data{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}(round_id: felt) -> (round: Round):
    # TODO: assert round_id fits in u32

    let (transmission: Transmission) = transmissions_.read(round_id)

    let round = Round(
        round_id=round_id,
        answer=transmission.answer,
        block_num=transmission.block_num,
        observation_timestamp=transmission.observation_timestamp,
        transmission_timestamp=transmission.transmission_timestamp,
    )
    return (round)
end

@view
func latest_round_data{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}() -> (round: Round):
    let (latest_round_id) = latest_aggregator_round_id_.read()
    let (transmission: Transmission) = transmissions_.read(latest_round_id)

    let round = Round(
        round_id=latest_round_id,
        answer=transmission.answer,
        block_num=transmission.block_num,
        observation_timestamp=transmission.observation_timestamp,
        transmission_timestamp=transmission.transmission_timestamp,
    )
    return (round)
end


# --- Set LINK Token

# --- Billing Access Controller

@storage_var
func billing_access_controller_() -> (access_controller: felt):
end

@external
func set_billing_access_controller{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}(access_controller: felt):
    Ownable_only_owner()

    let (old_controller) = billing_access_controller_.read()
    if access_controller != old_controller:
        billing_access_controller_.write(access_controller)
        # TODO: emit event

        tempvar syscall_ptr = syscall_ptr
        tempvar range_check_ptr = range_check_ptr
        tempvar pedersen_ptr = pedersen_ptr
    else:
        tempvar syscall_ptr = syscall_ptr
        tempvar range_check_ptr = range_check_ptr
        tempvar pedersen_ptr = pedersen_ptr
    end

    return ()
end

@view
func billing_access_controller{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}() -> (access_controller: felt):
    let (access_controller) = billing_access_controller_.read()
    return (access_controller)
end

# --- Billing Config

struct Billing:
    # TODO: use a single felt via (observation_payment, transmission_payment) = split_felt()?
    member observation_payment_gjuels : felt
    member transmission_payment_gjuels : felt
end

@storage_var
func billing_() -> (config: Billing):
end

@view
func billing{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}() -> (config: Billing):
    let (config: Billing) = billing_.read()
    return (config)
end

@external
func set_billing{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}(config: Billing):
    # TODO: use billing access controller too
    Ownable_only_owner()

    # Pay out oracles using existing settings for rounds up to now
    pay_oracles()

    # TODO: check payment value ranges within bounds (u32?)

    billing_.write(config)

    # TODO: emit event

    return ()
end

# --- Payments and Withdrawals

@external
func withdraw_payment{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}(transmitter: felt):
    # TODO: assert caller == receiver
    pay_oracle(transmitter)
    return ()
end

@external
func owed_payment{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}(transmitter: felt) -> (amount: felt):
    let (index) = transmitters_.read(transmitter)

    if index == 0:
        # let zero = Uint256(0, 0)
        return (0)
    end

    let (billing: Billing) = billing_.read()

    let (latest_round_id) = latest_aggregator_round_id_.read()
    let (from_round_id) = reward_from_aggregator_round_id_.read(transmitter)
    let rounds = latest_round_id - from_round_id

    # TODO: convert to juels
    let amount = rounds * billing.observation_payment_gjuels
    # TODO: add transmitter.payment

    # TODO: is felt enough here? or use Uint256
    return (amount)
end

@event
func oracle_paid(
    transmitter: felt,
    payee: felt,
    amount: Uint256,
    link_token: felt,
):
end

func pay_oracle{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}(transmitter: felt):
    alloc_locals

    let (index) = transmitters_.read(transmitter)

    if index == 0:
        return ()
    end

    let (amount_: felt) = owed_payment(transmitter)
    assert_nn(amount_)

    # if zero, fastpath return to avoid empty transfers
    let (not_zero) = is_not_zero(amount_)
    if not_zero == FALSE:
        return ()
    end

    let (amount: Uint256) = felt_to_uint256(amount_)
    let (payee) = payees_.read(transmitter)

    let (link_token) = link_token_.read()

    # TODO: do something with the return value?
    IERC20.transfer(
        contract_address=link_token,
        recipient=payee,
        amount=amount,
    )

    # TODO: emit event

    return ()
end

func pay_oracles{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}():
    # TODO: share link_token & last_round_id between pay_oracle calls

    # TODO: iter over transmitters_list_ and call pay_oracle
    return ()
end

# --- Transmitter Payment

func calculate_reimbursement() -> (amount: Uint256):
    # TODO:
    let amount = Uint256(0,0)
    return (amount)
end

# --- Payee Management

@storage_var
func payees_(transmitter: felt) -> (payment_address: felt):
end

@storage_var
func proposed_payees_(transmitter: felt) -> (payment_address: felt):
end

struct PayeeConfig:
    member transmitter: felt
    member payee: felt
end

@external
func set_payees{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr
}(payees_len: felt, payees: PayeeConfig*):
    Ownable_only_owner()

    # TODO:

    return()
end

@external
func transfer_payeeship{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}(transmitter: felt, proposed: felt):
    let (caller) = get_caller_address()
    let (payee) = payees_.read(transmitter)
    with_attr error_message("only current payee can update"):
        assert caller = payee
    end
    with_attr error_message("cannot transfer to self"):
        assert_not_equal(caller, proposed)
    end

    proposed_payees_.write(transmitter, proposed)

    # TODO: emit event

    return ()
end

@external
func accept_payeeship{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}(transmitter: felt):
    let (caller) = get_caller_address()
    let (proposed) = proposed_payees_.read(transmitter)
    with_attr error_message("only proposed payee can accept"):
        assert caller = proposed
    end

    payees_.write(transmitter, caller)
    proposed_payees_.write(transmitter, 0)

    # TODO: emit event

    return ()
end
