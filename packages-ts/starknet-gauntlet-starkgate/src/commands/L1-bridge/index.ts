import Deploy from './deploy'
import Deposit from './deposit'
import Withdraw from './withdraw'
import SetL2Bridge from './setL2Bridge'
import SetMaxTotalBalance from './setMaxTotalBalance'
import SetMaxDeposit from './setMaxDeposit'
import ProxyDeploy from './proxy/deploy'

export default [
  Deploy,
  SetL2Bridge,
  SetMaxTotalBalance,
  SetMaxDeposit,
  Deposit,
  Withdraw,
  ProxyDeploy,
]
