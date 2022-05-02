%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin, SignatureBuiltin
from starkware.cairo.common.hash import hash2
from starkware.cairo.common.hash_state import (
    hash_init, hash_finalize, hash_update, hash_update_single
)
from starkware.cairo.common.signature import verify_ecdsa_signature
from starkware.cairo.common.math import assert_not_zero, assert_not_equal, assert_lt, assert_nn_le, assert_nn, assert_in_range, unsigned_div_rem
# from starkware.cairo.common.bool import TRUE, FALSE

from starkware.starknet.common.syscalls import (
    get_caller_address,
    get_contract_address,
    get_block_timestamp,
    get_block_number
)

from openzeppelin.utils.constants import UINT8_MAX

from contracts.ownable import (
    Ownable_initializer,
    Ownable_only_owner,
    Ownable_get_owner,
    Ownable_transfer_ownership
)

# ---

const MAX_ORACLES = 31

@storage_var
func link_token_() -> (token: felt):
end

@storage_var
func billing_access_controller_() -> (access_controller: felt):
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
func description_() -> (decimals: felt):
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

struct Oracle:
  # TODO: payment amount, from_round_id
end

@storage_var
func oracles_len_() -> (len: felt):
end

@storage_var
func oracles_(index: felt) -> (oracle: Oracle):
end

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

# ---

@view
func owner{pedersen_ptr : HashBuiltin*, syscall_ptr : felt*, range_check_ptr}() -> (owner : felt):
    let (owner) = Ownable_get_owner()
    return (owner=owner)
end

# ---

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
    reimbursement: felt,
):
end

# ---

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
 
    # TODO: pay out existing oracles

    # remove old signers/transmitters
    let (len) = oracles_len_.read()
    remove_oracles(len)

    # add new oracles (also sets oracle_len_)
    add_oracles(oracles, 0, oracles_len)

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
    # TODO: validate signers unique
    verify_signatures(msg, signatures, signatures_len)

    # report():

    assert_nn_le(observations_len, MAX_ORACLES) # len <= MAX_ORACLES
    assert_lt(f, observations_len) # f < len

    latest_epoch_and_round_.write(report_context.epoch_and_round)

    let (median_idx : felt, _) = unsigned_div_rem(observations_len, 2)
    let median = observations[median_idx]

    # Validate median in min-max range
    let (answer_range) = answer_range_.read()
    assert_in_range(median, answer_range[0], answer_range[1])

    let (round_id) = latest_aggregator_round_id_.read()
    let round_id = round_id + 1
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

    # TODO: validate via validator

    # TODO: calculate reimbursement
    let reimbursement = 0

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

# ---

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
}(oracles: OracleConfig*, index: felt, len: felt):
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

    return add_oracles(oracles + OracleConfig.SIZE, index, len - 1)
end
# ---

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
    alloc_locals

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

func hash_report{
    pedersen_ptr : HashBuiltin*,
}(
    report_context: ReportContext,
    observers: felt,
    observations_len: felt,
    observations: felt*,
) -> (hash: felt):
    alloc_locals

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
    range_check_ptr
}(
    msg: felt,
    signatures: Signature*,
    signatures_len: felt
):
    alloc_locals
 
    if signatures_len == 0:
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

    return verify_signatures(
        msg,
        signatures + Signature.SIZE,
        signatures_len - 1
    )
end
