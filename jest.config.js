// TODO: it doesn't seem this file is actually used
module.exports = {
  rootDir: '.',
  projects: [
    {
      displayName: 'starknet-gauntlet',
      preset: 'ts-jest',
      testEnvironment: 'node',
      testMatch: ['<rootDir>/packages-ts/starknet-gauntlet/**/*.test.ts'],
      globals: {
        'ts-jest': {
          tsconfig: '<rootDir>/packages-ts/starknet-gauntlet/tsconfig.json',
        },
      },
    },
    {
      displayName: 'starknet-gauntlet-example',
      preset: 'ts-jest',
      testEnvironment: 'node',
      testMatch: ['<rootDir>/packages-ts/starknet-gauntlet-example/**/*.test.ts'],
      globals: {
        'ts-jest': {
          tsconfig: '<rootDir>/packages-ts/starknet-gauntlet-example/tsconfig.json',
        },
      },
    },
    {
      displayName: 'starknet-gauntlet-ocr2',
      preset: 'ts-jest',
      testEnvironment: 'node',
      testMatch: ['<rootDir>/packages-ts/starknet-gauntlet-ocr2/**/*.test.ts'],
      globals: {
        'ts-jest': {
          tsconfig: '<rootDir>/packages-ts/starknet-gauntlet-ocr2/tsconfig.json',
        },
      },
    },
    {
      displayName: 'starknet-gauntlet-oz',
      preset: 'ts-jest',
      testEnvironment: 'node',
      testMatch: ['<rootDir>/packages-ts/starknet-gauntlet-oz/**/*.test.ts'],
      globals: {
        'ts-jest': {
          tsconfig: '<rootDir>/packages-ts/starknet-gauntlet-oz/tsconfig.json',
        },
      },
    },
    {
      displayName: 'starknet-gauntlet-multisig',
      preset: 'ts-jest',
      testEnvironment: 'node',
      testMatch: ['<rootDir>/packages-ts/starknet-gauntlet-multisig/**/*.test.ts'],
      globals: {
        'ts-jest': {
          tsconfig: '<rootDir>/packages-ts/starknet-gauntlet-multisig/tsconfig.json',
        },
      },
    },
    {
      displayName: 'starknet-gauntlet-starkgate',
      preset: 'ts-jest',
      testEnvironment: 'node',
      testMatch: ['<rootDir>/packages-ts/starknet-gauntlet-starkgate/**/*.test.ts'],
      globals: {
        'ts-jest': {
          tsconfig: '<rootDir>/packages-ts/starknet-gauntlet-starkgate/tsconfig.json',
        },
      },
    },
    {
      displayName: 'starknet-gauntlet-argent',
      preset: 'ts-jest',
      testEnvironment: 'node',
      testMatch: ['<rootDir>/packages-ts/starknet-gauntlet-argent/**/*.test.ts'],
      globals: {
        'ts-jest': {
          tsconfig: '<rootDir>/packages-ts/starknet-gauntlet-argent/tsconfig.json',
        },
      },
    },
  ],
}
