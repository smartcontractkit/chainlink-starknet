import { LEDGER_PATH_REGEX, DEFAULT_LEDGER_PATH } from '../src/'

describe('Ledger Input Path', () => {
  const wrongPathFormat1 = 'hola!'
  const wrongPathFormat2 = "m/2645'/579218131'/0'/16'/67'/123'"
  const wrongPathStartWith1 = "m/4417'/667845903'/894929996'/0'"
  const wrongPathStartWith2 = "  path 2645'/579218131'/894929996'/0'"
  const pathWithSpaces = "m  / 2645'/ 579218131'  / 0 '/0 '"

  it("Doesn't validate wrong format", async () => {
    let res = LEDGER_PATH_REGEX.exec(wrongPathFormat1)
    expect(res).toBeFalsy()

    res = LEDGER_PATH_REGEX.exec(wrongPathFormat2)
    expect(res).toBeFalsy()
  })

  it("Doesn't validate not starting with m/2645'/579218131'/", async () => {
    let res = LEDGER_PATH_REGEX.exec(wrongPathStartWith1)
    expect(res).toBeFalsy()

    res = LEDGER_PATH_REGEX.exec(wrongPathStartWith2)
    expect(res).toBeFalsy()
  })

  it('Validates path with spaces', async () => {
    const res = LEDGER_PATH_REGEX.exec(pathWithSpaces)
    expect(res).toBeTruthy()
  })

  it('Validates default path', async () => {
    const res = LEDGER_PATH_REGEX.exec(DEFAULT_LEDGER_PATH)
    expect(res).toBeTruthy()
  })
})
