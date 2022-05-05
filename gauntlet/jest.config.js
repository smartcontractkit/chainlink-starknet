module.exports = {
  rootDir: '.',
  projects: [
    {
      displayName: 'gauntlet-starknet',
      preset: 'ts-jest',
      testEnvironment: 'node',
      testMatch: ['<rootDir>/packages/gauntlet-starknet/**/*.test.ts'],
      globals: {
        'ts-jest': {
          tsconfig: '<rootDir>/packages/gauntlet-starknet/tsconfig.json',
        },
      },
    },
    {
      displayName: 'gauntlet-starknet-ocr2',
      preset: 'ts-jest',
      testEnvironment: 'node',
      testMatch: ['<rootDir>/packages/gauntlet-starknet-ocr2/**/*.test.ts'],
      globals: {
        'ts-jest': {
          tsconfig: '<rootDir>/packages/gauntlet-starknet-ocr2/tsconfig.json',
        },
      },
    },
  ],
}
