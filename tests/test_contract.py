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
    # for oracle in oracles:
    #     await starknet.deploy(
    #         contract_path("account.cairo"),
    #         constructor_calldata=[oracle.public_key]
    #     )
    
    # Call set_config
    
    f = 1
    # onchain_config = []
    onchain_config = 1
    offchain_config_version = 2
    offchain_config = [1]
    
    # TODO: need to call via owner
    await contract.set_config(
        oracles=[(
            oracle['signer'].public_key,
            oracle['transmitter'].public_key
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

    oracle = oracles[0]
    # transmitter = Signer(123456789987654321)
    
    account = await starknet.deploy(
        contract_path("account.cairo"),
        constructor_calldata=[oracle['transmitter'].public_key]
    )
    
    report_context = bytes([0x0])
    observers = bytes([0x0])
    observations = [99]
    
    msg = compute_hash_on_elements([
        int.from_bytes(report_context, "big"),
        int.from_bytes(observers, "big"),
        len(observations),
        *observations,
    ])
    
    # Sign with a single oracle
    sig_r, sig_s = sign(msg_hash=msg, priv_key=oracle['signer'].private_key)
    
    signatures = [
        sig_r, # r
        sig_s, # s
        oracle['signer'].public_key  # public_key
    ]
    
    calldata = [
        int.from_bytes(report_context, "big"),
        int.from_bytes(observers, "big"),
        len(observations),
        *observations,
        1, # len signatures
        *signatures # TODO: how to convert objects to calldata? using array for now
    ]
    
    await oracle['transmitter'].send_transaction(
        account,
        contract.contract_address,
        'transmit',
        calldata
    )
    
