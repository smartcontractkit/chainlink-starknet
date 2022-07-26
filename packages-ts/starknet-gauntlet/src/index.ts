// Node.js 18 + Jest 28 are supposed to support fetch but I can't get it working.
// Copying starknet.js workaround: https://github.com/0xs34n/starknet.js/commit/83be37a9e3328a44abd9583b8167c3cb8d882790
import fetch from 'cross-fetch'
if (!global.fetch) {
  global.fetch = fetch
}

export * from './commands/base'
export * from './dependencies'
export * from './provider'
export * from './wallet'
export * from './events'
export * from './utils'
