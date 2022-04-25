"""contract.cairo test file."""
import os

import pytest
from starkware.starknet.testing.starknet import Starknet

from utils import (
    Signer, to_uint, add_uint, sub_uint, str_to_felt, MAX_UINT256, ZERO_ADDRESS, INVALID_UINT256,
    get_contract_def, contract_path, cached_contract, assert_revert, assert_event_emitted, uint
)

signer = Signer(123456789987654321)

# The path to the contract source code.
CONTRACT_FILE = os.path.join("contracts", "contract.cairo")

@pytest.fixture(scope='module')
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

# The testing library uses python's asyncio. So the following
# decorator and the ``async`` keyword are needed.
@pytest.mark.asyncio
async def test_increase_balance(token_factory):
    """Test increase_balance method."""
    starknet, token, owner = token_factory    

    # Deploy the contract.
    contract = await starknet.deploy(
        source=CONTRACT_FILE,
    )

    # Invoke increase_balance() twice.
    await contract.increase_balance(amount=10).invoke()
    await contract.increase_balance(amount=20).invoke()

    # Check the result of get_balance().
    execution_info = await contract.get_balance().call()
    assert execution_info.result == (30,)
