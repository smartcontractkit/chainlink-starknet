import { LedgerSigner } from "./src/index";
import { ec} from 'starknet'

async function main() {
  console.log("Initializing ledger client")
  const signer = await LedgerSigner.create();
  const pubkey = await signer.getPubKey();
  console.log(`Public key: ${pubkey}`)

  const hash = 'c465dd6b1bbffdb05442eb17f5ca38ad1aa78a6f56bf4415bdee219114a47';
  console.log(`Signing hash ${hash} via ledger...`)
  const signature = await signer.signRaw(hash);
  console.log(`Signature: {signature}`)

  const matches = ec.starkCurve.verify(signature.toString(), hash, pubkey);
  console.log(`Signature verifies: ${matches}`)
}
  

main()
