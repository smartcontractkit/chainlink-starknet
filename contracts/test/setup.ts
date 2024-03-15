import * as path from 'node:path'
import * as fs from 'node:fs'

function findCommonPrefix(path1: string, path2: string): string {
  const segments1 = path1.split(path.sep)
  const segments2 = path2.split(path.sep)

  const minLength = Math.min(segments1.length, segments2.length)
  const commonSegments = []

  for (let i = 0; i < minLength; i++) {
    if (segments1[i] === segments2[i]) {
      commonSegments.push(segments1[i])
    } else {
      break
    }
  }

  return commonSegments.join(path.sep)
}

function toCamelCase(str: string): string {
  return str
    .replace(/_([a-z])/g, (_, letter) => letter.toUpperCase())
    .replace(/-([a-z])/g, (_, letter) => letter.toUpperCase())
    .replace(/(^| )([a-z])/g, (_, __, letter) => letter.toUpperCase())
    .replace(/ /g, '')
}

function findCairoFiles(dir: string): string[] {
  const entries = fs.readdirSync(dir, { withFileTypes: true })
  const filePaths = entries.flatMap((entry) => {
    const entryPath = path.join(dir, entry.name)
    if (entry.isDirectory()) {
      return findCairoFiles(entryPath)
    } else if (entry.isFile() && entry.name.toLowerCase().endsWith('.cairo')) {
      return [entryPath]
    } else {
      return []
    }
  })
  return filePaths
}

export function prepareHardhatArtifacts() {
  const hre = require('hardhat')

  const src = hre.config.paths.starknetSources
  const target = hre.config.paths.starknetArtifacts
  if (!src || !target) {
    throw new Error('Missing starknet path config')
  }

  const root = findCommonPrefix(src, target)

  console.log('Cleaning and regenerating hardhat file structure..')
  const generatedPath = path.join(target, src.slice(root.length))
  if (fs.existsSync(generatedPath)) {
    fs.rmSync(generatedPath, { recursive: true })
  }

  const cairoFiles = findCairoFiles(src)
  for (const cairoFile of cairoFiles) {
    const relativePath = cairoFile
    const filename = path.basename(relativePath, '.cairo')

    const camelCaseFilename = toCamelCase(filename)

    const sierraFile = `${target}/chainlink_${camelCaseFilename}.sierra.json`
    const casmFile = `${target}/chainlink_${camelCaseFilename}.casm.json`

    if (!fs.existsSync(sierraFile) || !fs.existsSync(casmFile)) {
      continue
    }

    const subdir = path.dirname(relativePath).slice(root.length)
    // Create the corresponding directory
    const targetSubdir = path.join(target, subdir, `${filename}.cairo`)
    fs.mkdirSync(targetSubdir, { recursive: true })

    // Copy the sierra and casm files. We need to copy instead of symlink
    // because hardhat-starknet-plugin does fs.lstatSync to check if the file
    // exists.
    fs.copyFileSync(sierraFile, `${targetSubdir}/${filename}.json`)
    fs.copyFileSync(casmFile, `${targetSubdir}/${filename}.casm`)

    // Parse and save the ABI JSON
    const sierraContent = JSON.parse(fs.readFileSync(sierraFile, 'utf8'))
    fs.writeFileSync(
      `${targetSubdir}/${filename}_abi.json`,
      JSON.stringify(sierraContent.abi, null, 2),
    )
  }
}
