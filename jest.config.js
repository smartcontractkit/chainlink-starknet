module.exports = {
  rootDir: '.',
  projects: [
    // {
    //   displayName: 'gauntlet-starknet',
    //   preset: 'ts-jest',
    //   testEnvironment: 'node',
    //   testMatch: ['<rootDir>/packages-ts/gauntlet-starknet/**/*.test.ts'],
    //   globals: {
    //     'ts-jest': {
    //       tsconfig: '<rootDir>/packages-ts/gauntlet-starknet/tsconfig.json',
    //     },
    //   },
    // },
    // {
    //   displayName: 'gauntlet-starknet-example',
    //   preset: 'ts-jest',
    //   testEnvironment: 'node',
    //   testMatch: ['<rootDir>/packages-ts/gauntlet-starknet-example/**/*.test.ts'],
    //   globals: {
    //     'ts-jest': {
    //       tsconfig: '<rootDir>/packages-ts/gauntlet-starknet-example/tsconfig.json',
    //     },
    //   },
    // },
    {
      displayName: 'gauntlet-starknet-ocr2',
      preset: 'ts-jest',
      testEnvironment: 'node',
      testMatch: ['<rootDir>/packages-ts/gauntlet-starknet-ocr2/**/*.test.ts'],
      globals: {
        'ts-jest': {
          tsconfig: '<rootDir>/packages-ts/gauntlet-starknet-ocr2/tsconfig.json',
        },
      },
    },
    // {
    //   displayName: 'gauntlet-starknet-oz',
    //   preset: 'ts-jest',
    //   testEnvironment: 'node',
    //   testMatch: ['<rootDir>/packages-ts/gauntlet-starknet-oz/**/*.test.ts'],
    //   globals: {
    //     'ts-jest': {
    //       tsconfig: '<rootDir>/packages-ts/gauntlet-starknet-oz/tsconfig.json',
    //     },
    //   },
    // },
    // {
    //   displayName: 'gauntlet-starknet-starkgate',
    //   preset: 'ts-jest',
    //   testEnvironment: 'node',
    //   testMatch: ['<rootDir>/packages-ts/gauntlet-starknet-starkgate/**/*.test.ts'],
    //   globals: {
    //     'ts-jest': {
    //       tsconfig: '<rootDir>/packages-ts/gauntlet-starknet-starkgate/tsconfig.json',
    //     },
    //   },
    // },
    // {
    //   displayName: 'gauntlet-starknet-argent',
    //   preset: 'ts-jest',
    //   testEnvironment: 'node',
    //   testMatch: ['<rootDir>/packages-ts/gauntlet-starknet-argent/**/*.test.ts'],
    //   globals: {
    //     'ts-jest': {
    //       tsconfig: '<rootDir>/packages-ts/gauntlet-starknet-argent/tsconfig.json',
    //     },
    //   },
    // },
  ],
}
