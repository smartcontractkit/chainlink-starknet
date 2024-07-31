import { shortString } from 'starknet'

const strings = shortString.splitLongString(
  '0x0000000000004163636f756e743a20696e76616c6964207369676e6174757265',
)
console.log(strings)
strings.forEach((s) => {
  shortString.decodeShortString(s)
})
