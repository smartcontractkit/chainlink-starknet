import Deploy from './deploy'
import Declare from './declare'
import Upgrade from './upgrade'
import ProposeOwnership from './proposeOwnership'
import AcceptOwnership from './acceptOwnership'

export const executeCommands = [Declare, Deploy, Upgrade, ProposeOwnership, AcceptOwnership]
