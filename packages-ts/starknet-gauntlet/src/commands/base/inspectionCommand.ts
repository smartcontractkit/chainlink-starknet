import BaseCommand from '@chainlink/gauntlet-core/dist/commands/internal/base'
import { CompiledContract, CompiledSierraCasm, Contract, hash } from 'starknet'
import { CommandCtor, Input } from '.'
import { InspectionDependencies } from '../../dependencies'
import { IStarknetProvider } from '../../provider'
import { CommandUX, makeCommandId } from './command'
import { LoadContractResult } from './executeCommand'

export interface InspectUserInput<UI, CompareInput> {
  input: UI
  toCompare?: CompareInput
}

// TODO: Temporary inspection report.
export interface InspectionReport<QueryResult> {
  data: QueryResult
  contract: string
  inspection: {
    id: string
    message: string
    resultType: 'success' | 'failed'
  }[]
}

export interface InspectCommandConfig<UI, CI, CompareInput, QueryResult> {
  ux: CommandUX
  // List of query functions to call
  queries: string[]
  makeUserInput?: (flags: any, args: string[]) => Promise<InspectUserInput<UI, CompareInput>>
  /**
   * Given the user input, translate to every contract input required for each query
   */
  makeContractInput?: (userInput: UI) => Promise<CI[]>
  /**
   * After doing every query, convert the results into the type we want (QueryResult) and if toCompare is given, match result into it
   */
  makeComparisionData: (
    provider: IStarknetProvider,
  ) => (
    results: any[],
    input: UI,
    contractAddress: string,
  ) => Promise<{
    toCompare: CompareInput
    result: QueryResult
  }>
  inspect?: (
    expected: InspectUserInput<UI, CompareInput>,
    data: {
      toCompare: CompareInput
      result: QueryResult
    },
  ) => {
    id: string
    message: string
    resultType: 'success' | 'failed'
  }[]
  loadContract: () => LoadContractResult
}

export interface InspectCommandInstance<QueryResult> {
  execute: () => Promise<{
    data: InspectionReport<QueryResult>
    responses: any[]
  }>
}

export const makeInspectionCommand = <UI, CI, CompareInput, QueryResult>(
  config: InspectCommandConfig<UI, CI, CompareInput, QueryResult>,
) => (deps: InspectionDependencies) => {
  const command: CommandCtor<InspectCommandInstance<QueryResult>> = class InspectionCommand
    extends BaseCommand
    implements InspectCommandInstance<QueryResult> {
    // Props
    provider: IStarknetProvider
    contractAddress: string

    input: Input<InspectUserInput<UI, CompareInput>, CI>

    contract: CompiledContract
    compiledContractHash?: string

    // UX
    static id = makeCommandId(config.ux.category, config.ux.function, config.ux.suffixes)
    static category = config.ux.category
    static examples = config.ux.examples

    static create = async (flags, args) => {
      const c = new InspectionCommand(flags, args)

      const env = deps.makeEnv(flags)

      c.provider = deps.makeProvider(env.providerUrl)
      c.contractAddress = args[0]

      c.input = await c.buildCommandInput(flags, args)
      const loadResult = config.loadContract()
      c.contract = loadResult.contract
      if (loadResult.casm) {
        c.compiledContractHash = hash.computeCompiledClassHash(loadResult.casm)
      }

      return c
    }

    buildCommandInput = async (
      flags,
      args,
    ): Promise<Input<InspectUserInput<UI, CompareInput>, CI>> => {
      const userInput = config.makeUserInput && (await config.makeUserInput(flags, args))
      const contractInput =
        config.makeContractInput && (await config.makeContractInput(userInput.input))

      return {
        user: userInput || {
          input: null,
          toCompare: null,
        },
        contract: contractInput || [],
      }
    }

    runQueries = async (functions: string[], contractInputs: CI | CI[]): Promise<any[]> => {
      const inputs = Array.isArray(contractInputs) ? contractInputs : [contractInputs]
      const contract = new Contract(this.contract.abi, this.contractAddress, this.provider.provider)
      const results = await Promise.all(
        functions.map((func, i) => {
          deps.logger.loading(`Fetching ${func} of contract ${this.contractAddress}...`)
          if (!inputs[i]) {
            return contract[func]() // workaround undefined argument inputs[i]
          }
          return contract[func](inputs[i])
        }),
      )
      return results
    }

    execute = async () => {
      const results = await this.runQueries(config.queries, this.input.contract)
      const data = await config.makeComparisionData(this.provider)(
        results,
        this.input.user.input,
        this.contractAddress,
      )
      const inspectionResults = config.inspect ? config.inspect(this.input.user, data) : []

      deps.logger.info('Inspection Results:')
      deps.logger.log(data.result)
      // TODO: Gauntlet core forces us to use Result type for every command. Update to choose the result if using Base Command
      return {
        data: {
          data: data.result,
          contract: this.contractAddress,
          inspection: inspectionResults,
        },
        responses: [],
      }
    }
  }

  return command
}
