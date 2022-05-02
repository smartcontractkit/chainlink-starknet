"""contract.cairo test file."""
import os

import pytest
import pytest_asyncio
from starkware.starknet.testing.starknet import Starknet
from starkware.crypto.signature.signature import (
    pedersen_hash, private_to_stark_key, sign)
from starkware.cairo.common.hash_state import compute_hash_on_elements

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

@pytest_asyncio.fixture(scope='module')
async def token_factory():
    # Create a new Starknet class that simulates the StarkNet system.
    starknet = await Starknet.empty()
    owner = await starknet.deploy(
        contract_path("account.cairo"),
        constructor_calldata=[signer.public_key]
    )

    token = await starknet.deploy(
        contract_path("token.cairo"),
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
#         source=contract_path("contract.cairo"),
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

@pytest.mark.asyncio
async def test_transmit(token_factory):
    """Test constructor method."""
    starknet, token, owner = token_factory    

    # Deploy the contract.
    contract = await starknet.deploy(
        source=contract_path("contract.cairo"),
        constructor_calldata=[
            owner.contract_address,
            token.contract_address,
            0,
            1000000000,
            0, # TODO: billing AC
            8, # decimals
            str_to_felt("ETH/BTC")
        ]
    )

    # Deploy an account for each oracle
    for oracle in oracles:
        oracle['account'] = await starknet.deploy(
            contract_path("account.cairo"),
            constructor_calldata=[oracle['transmitter'].public_key]
        )
    
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

    print(f"digest = {digest}")

    oracle = oracles[0]

    # TODO:
    epoch_and_round = 1
    extra_hash = 1
    report_context = [digest, epoch_and_round, extra_hash]
    # int.from_bytes(report_context, "big"),

    observers = bytes([i for i in range(len(oracles))])
    observations = [99 for _ in range(len(oracles))]
    
    raw_report = [
        int.from_bytes(observers, "big"),
        len(observations),
        *observations,
    ]
    
    msg = compute_hash_on_elements([
        *report_context,
        *raw_report
    ])
    
    n = f + 1

    signatures = []
    
    for oracle in oracles[:n]:
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
        *signatures # TODO: how to convert objects to calldata? using array for now
    ]
    
    print(calldata)
    
    await oracle['transmitter'].send_transaction(
        oracle['account'],
        contract.contract_address,
        'transmit',
        calldata
    )
    
