# Chainlink ERC677

## starkgate_token.cairo

Similare to the starkgate ERC20 but with some added functionality from ERC677.

- transferAndCall -> Transfers tokens to receiver, via ERC20's transfer(address,uint256) function. It then logs an event Transfer(address,address,uint256,bytes). Once the transfer has succeeded and the event is logged, the token calls onTokenTransfer(selector, data_len, data) on the receiver with the function's selector, and all the parameters required by the function that you want to call next.

## linkReceiver.cairo

- onTokenTransfer -> The function is added to contracts enabling them to react to receiving tokens within a single transaction. The selector parameter is the selector of the function that you want to call.
  The data paramater is all the parameters required by the function that you want to call through the selector.
