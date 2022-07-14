import { getSelectorFromName } from 'starknet/dist/utils/hash'

function main() {
  const setSelector = getSelectorFromName('updateStatus')
  console.log(BigInt(setSelector).toString(10))
}
main()
