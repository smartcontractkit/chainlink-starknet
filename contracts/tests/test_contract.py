"""aggregator.cairo test file."""
import os

import asyncio
import pytest
import pytest_asyncio
from starkware.starknet.testing.starknet import Starknet
from starkware.crypto.signature.signature import (
    pedersen_hash, private_to_stark_key, sign)
from starkware.cairo.common.hash_state import compute_hash_on_elements
from starkware.starknet.utils.api_utils import cast_to_felts

from utils import (
    Signer, to_uint, add_uint, sub_uint, str_to_felt, MAX_UINT256, ZERO_ADDRESS, INVALID_UINT256,
    get_contract_def, contract_path, cached_contract, assert_revert, assert_event_emitted, uint
)

signer = Signer(999654321123456789)

oracles = [
    { 'signer': Signer(123456789987654321), 'transmitter': Signer(987654321123456789) },
    { 'signer': Signer(123456789987654322), 'transmitter': Signer(987654321123456788) },
    { 'signer': Signer(123456789987654323), 'transmitter': Signer(987654321123456787) },
    { 'signer': Signer(123456789987654324), 'transmitter': Signer(987654321123456786) },
]


@pytest.fixture(scope='module')
def contract_defs():
    aggregator_def = get_contract_def('aggregator.cairo')
    account_def = get_contract_def('account.cairo')
    erc20_def = get_contract_def('token.cairo')
    return aggregator_def, account_def, erc20_def

@pytest_asyncio.fixture(scope='module')
async def token_factory(contract_defs):
    _aggregator_def, account_def, erc20_def = contract_defs

    # Create a new Starknet class that simulates the StarkNet system.
    starknet = await Starknet.empty()

    owner = await starknet.deploy(
        contract_def=account_def,
        constructor_calldata=[signer.public_key]
    )

    token = await starknet.deploy(
        contract_def=erc20_def,
        constructor_calldata=[
            str_to_felt("LINK Token"),
            str_to_felt("LINK"),
            18,
            *uint(1000),
            owner.contract_address,
            owner.contract_address
        ]
    )
    return starknet, token, owner

# @pytest.mark.asyncio
# async def test_ownership(token_factory):
#     """Test constructor method."""
#     starknet, token, owner = token_factory    

#     # Deploy the contract.
#     contract = await starknet.deploy(
#         source=contract_path("aggregator.cairo"),
#         constructor_calldata=[
#             owner.contract_address,
#             token.contract_address,
#             0,
#             1000000000,
#             0, # TODO: billing AC
#             8, # decimals
#             str_to_felt("ETH/BTC")
#         ]
#     )

#     # # Invoke increase_balance() twice.
#     # await contract.increase_balance(amount=10).invoke()
#     # await contract.increase_balance(amount=20).invoke()

#     # Check the result of owner().
#     execution_info = await contract.owner().call()
#     assert execution_info.result == (owner.contract_address,)

# TODO: module scope won't work, need to wrap with a state copy
@pytest_asyncio.fixture(scope='module')
async def setup(token_factory, contract_defs):
    starknet, token, owner = token_factory
    aggregator_def, account_def, _erc20_def = contract_defs

    # Deploy the contract.
    min_answer = -10
    max_answer = 1000000000

    contract = await starknet.deploy(
        contract_def=aggregator_def,
        constructor_calldata=cast_to_felts([
            owner.contract_address,
            token.contract_address,
            *cast_to_felts(values=[
                min_answer,
                max_answer
            ]),
            0, # TODO: billing AC
            8, # decimals
            str_to_felt("ETH/BTC")
        ])
    )

    # Deploy an account for each oracle
    accounts = await asyncio.gather(*[
        starknet.deploy(
            contract_def=account_def,
            constructor_calldata=[oracle['transmitter'].public_key]
        ) for oracle in oracles
    ])
    for i in range(len(accounts)):
        oracles[i]['account'] = accounts[i]
    
    # Call set_config

    f = 1
    # onchain_config = []
    onchain_config = 1
    offchain_config_version = 2
    offchain_config = [1]
    
    # TODO: need to call via owner
    execution_info = await contract.set_config(
        oracles=[(
            oracle['signer'].public_key,
            oracle['account'].contract_address
        ) for oracle in oracles],
        # TODO: dict was supposed to be ok but it asks for a tuple
        # oracles=[{
        #     'signer': oracle['signer'].public_key,
        #     'transmitter': oracle['transmitter'].public_key
        # } for oracle in oracles],
        f=f,
        onchain_config=onchain_config,
        offchain_config_version=2,
        offchain_config=offchain_config
    ).invoke()

    digest = execution_info.result.digest

    return {
        "starknet": starknet,
        "token": token,
        "owner": owner,
        "contract": contract,
        "f": f,
        "digest": digest,
        "oracles": oracles
    }

@pytest.fixture
def aggregator_factory(contract_defs, setup):
    aggregator_def, account_def, erc20_def = contract_defs
    env = setup
    _state = env["starknet"].state.copy()
    token = cached_contract(_state, erc20_def, env["token"])
    owner = cached_contract(_state, account_def, env["owner"])
    contract = cached_contract(_state, aggregator_def, env["contract"])
    
    # TODO: need to replace all oracles with cache too?

    return {
        "token": token,
        "owner": owner,
        "contract": contract,
        "f": env["f"],
        "digest": env["digest"],
        "oracles": env["oracles"]
    }


@pytest.mark.asyncio
async def test_transmit(aggregator_factory):
    """Test transmit method."""
    env = aggregator_factory
    print(f"digest = {env['digest']}")

    oracle = env["oracles"][0]

    n = env["f"] + 1
    
    def transmit(
        epoch_and_round, # TODO: split into two values
        answer
    ):
        # TODO:
        observation_timestamp = 1
        extra_hash = 1
        juels_per_fee_coin = 1
        report_context = [env["digest"], epoch_and_round, extra_hash]
        # int.from_bytes(report_context, "big"),

        l = len(env["oracles"])
        observers = bytes([i for i in range(l)])
        observations = [answer for _ in range(l)]
    

        raw_report = [
            observation_timestamp,
            int.from_bytes(observers, "big"),
            len(observations),
            *cast_to_felts(observations), # convert negative numbers to valid felts
            juels_per_fee_coin,
        ]
    
        msg = compute_hash_on_elements([
            *report_context,
            *raw_report
        ])

        signatures = []
    
        # TODO: test with duplicate signers
        # for o in oracles[:n]:
        #     oracle = oracles[0]

        for oracle in env["oracles"][:n]:
            # Sign with a single oracle
            sig_r, sig_s = sign(msg_hash=msg, priv_key=oracle['signer'].private_key)
    
            signature = [
                sig_r, # r
                sig_s, # s
                oracle['signer'].public_key  # public_key
            ]
            signatures.extend(signature)

        calldata = [
            *report_context,
            *raw_report,
            n, # len signatures
            *signatures
        ]
    
        print(calldata)
    
        return oracle['transmitter'].send_transaction(
            oracle['account'],
            env["contract"].contract_address,
            'transmit',
            calldata
        )

    await transmit(epoch_and_round=1, answer=99)
    # TODO: test latest_round_data, round_data
    await transmit(epoch_and_round=2, answer=-1)
    
@pytest.mark.asyncio
async def test_payees(aggregator_factory):
    """Test payee related functionality."""
    env = aggregator_factory
    
    payees = []

    for oracle in env["oracles"]:
        payees.extend([
            oracle['transmitter'].public_key,
            oracle['transmitter'].public_key # reusing transmitter accounts as payees for simplicity
        ])

    calldata = [
        len(env["oracles"]),
        *payees
    ]

    # should succeed because all payees are zero
    execution_info = await signer.send_transaction(
        env['owner'],
        env["contract"].contract_address,
        'set_payees',
        calldata
    )
    
    # should succeed because all payees equal current payees
    execution_info = await signer.send_transaction(
        env['owner'],
        env["contract"].contract_address,
        'set_payees',
        calldata
    )

    # can't transfer to self
    assert_revert(signer.send_transaction(
        env['owner'],
        env["contract"].contract_address,
        'transfer_payeeship',
        [transmitter, proposed]
    ))
    # only payee can transfer
    # successful transfer
    # only proposed payee can accept
    # successful accept

