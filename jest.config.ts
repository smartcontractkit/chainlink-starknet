// jest.config.ts
import type { Config } from '@jest/types'

// Prepares 'config.projects' entry for a Jest TS project under '<rootDir>/packages-ts'
const projectConfig = (name: string) => ({
  displayName: name,
  testMatch: [`<rootDir>/packages-ts/${name}/**/*.test.ts`],
  transform: {
    '^.+\\.(ts|tsx)$': 'ts-jest',
  },
  globals: {
    'ts-jest': {
      tsconfig: `<rootDir>/packages-ts/${name}/tsconfig.json`,
    },
  },
})

const config: Config.InitialOptions = {
  rootDir: '.',
  preset: 'ts-jest',
  testEnvironment: 'node',
  verbose: true,
  automock: true,
  testPathIgnorePatterns: ['dist/', 'node_modules/'],
  projects: [
    projectConfig('starknet-gauntlet'),
    projectConfig('starknet-gauntlet-argent'),
    projectConfig('starknet-gauntlet-cli'),
    projectConfig('starknet-gauntlet-example'),
    projectConfig('starknet-gauntlet-multisig'),
    projectConfig('starknet-gauntlet-ocr2'),
    projectConfig('starknet-gauntlet-oz'),
    projectConfig('starknet-gauntlet-token'),
    projectConfig('starknet-gauntlet-emergency-protocol'),
    projectConfig('starknet-gauntlet-ledger'),
  ],
}
export default config
