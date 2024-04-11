import { LedgerSigner } from './src/index'
import { encode, Signature, ec, Signer, WeierstrassSignatureType } from 'starknet'

async function main() {
  const privateKey = '0x1234567890987654321'
  const starknetPublicKey = ec.starkCurve.getStarkKey(privateKey)
  const fullPublicKey = encode.addHexPrefix(
    encode.buf2hex(ec.starkCurve.getPublicKey(privateKey, false)),
  )
  console.log(fullPublicKey)

  console.log('Initializing ledger client')
  const signer = await LedgerSigner.create()
  const pubkey = await signer.getPubKey()
  console.log(`Public key: ${pubkey}`)

  const hash = 'c465dd6b1bbffdb05442eb17f5ca38ad1aa78a6f56bf4415bdee219114a47'
  console.log(`Signing hash ${hash} via ledger...`)
  const sig = await signer.signRaw(hash)
  const signature = sig as WeierstrassSignatureType
  console.log(`Signature: ${signature.toCompactHex()} r=${signature.r} s=${signature.s}`)

  const matches = ec.starkCurve.verify(signature, hash, `0x${pubkey}`)
  console.log(`Signature verifies: ${matches}`)
}

main()
