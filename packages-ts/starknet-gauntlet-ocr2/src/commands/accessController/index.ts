import Deploy from './deploy'
import Declare from './declare'
import Upgrade from './upgrade'
import TransferOwnership from './transferOwnership'
import AcceptOwnership from './acceptOwnership'

export const executeCommands = [Declare, Deploy, Upgrade, TransferOwnership, AcceptOwnership]
