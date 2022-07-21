import { getSelectorFromName } from 'starknet/dist/utils/hash'

function main() {
  const setSelector = getSelectorFromName('update_status')
  console.log(BigInt(setSelector).toString(10))
}
main()
