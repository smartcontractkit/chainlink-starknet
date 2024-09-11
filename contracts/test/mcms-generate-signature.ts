import { Wallet, utils } from 'ethers'

const SIGNER_ADDRESS_1 = '0x13Cf92228941e27eBce80634Eba36F992eCB148A'
const SIGNER_PRIVATE_KEY_1 = '0xf366414c9042ec470a8d92e43418cbf62caabc2bbc67e82bd530958e7fcaa688'
const SIGNER_ADDRESS_2 = '0xDa09C953823E1F60916E85faD44bF99A7DACa267'
const SIGNER_PRIVATE_KEY_2 = '0xed10b7a09dd0418ab35b752caffb70ee50bbe1fe25a2ebe8bba8363201d48527'

const toBytes = (message: string): Uint8Array => {
  const numBytes = message.length / 2
  const bytes = new Uint8Array(numBytes)

  for (let i = 0; i < numBytes; i++) {
    const byteValue = parseInt(message.slice(2 * i, 2 * i + 2), 16)
    bytes[i] = byteValue
  }

  return bytes
}

const generateSignature = async (merkleRoot: string) => {
  const signer1 = new Wallet(new utils.SigningKey(SIGNER_PRIVATE_KEY_1))
  const signer2 = new Wallet(new utils.SigningKey(SIGNER_PRIVATE_KEY_2))

  const signers = [signer1, signer2]

  const merkleRootHex = BigInt(merkleRoot).toString(16)

  const msgToSign = toBytes(merkleRootHex)

  for (let i = 0; i < signers.length; i++) {
    const signature = utils.splitSignature(await signers[i].signMessage(msgToSign))

    const highR = '0x' + signature.r.slice(2, 34); // First 32 characters (high u128)
    const lowR = '0x' + signature.r.slice(34);    // Last 32 characters (low u128)

    const highS = '0x' + signature.s.slice(2, 34); // First 32 characters (high u128)
    const lowS = '0x' + signature.s.slice(34);    // Last 32 characters (low u128)

    console.log(`
      signature ${i + 1}: 
        r: 
          high: ${highR}, low: ${lowR}
        s:
          high: ${highS}, low: ${lowS}
        v:
          ${signature.v}
    `)
  }
}


// run scarb test -e chainlink::tests::test_mcms::test_set_root::test_set_root_success and observe standard out to get the merkle root
// then execute `yarn ts-node test/mcms-generate-signature.ts <output>`
const main = async () => {

  await generateSignature(process.argv[2])
}

main()
