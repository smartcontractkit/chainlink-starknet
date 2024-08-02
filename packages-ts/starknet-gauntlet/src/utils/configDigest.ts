import fs from 'fs'
import { CONTRACT_TYPES, getRDD } from '../rdd'
import { Dependencies } from '../dependencies'

export const tryToWriteLastConfigDigestToRDD = async (
  deps: Pick<Dependencies, 'logger' | 'prompt'>,
  rddPath: string,
  contractAddr: string,
  configDigest: string,
) => {
  deps.logger.info(`lastConfigDigest to save in RDD: ${configDigest}`)
  if (rddPath) {
    const rdd = getRDD(rddPath)
    // set updated lastConfigDigest
    rdd[CONTRACT_TYPES.AGGREGATOR][contractAddr]['config']['lastConfigDigest'] = configDigest
    fs.writeFileSync(rddPath, JSON.stringify(rdd, null, 2))
    deps.logger.success(
      `RDD file ${rddPath} has been updated! You must reformat RDD by running ./bin/degenerate and ./bin/generate in that exact order`,
    )
  } else {
    deps.logger.warn(
      `No RDD file was inputted, you must manually update lastConfigDigest in RDD yourself`,
    )
  }
}
