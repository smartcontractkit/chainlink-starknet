export interface CommandCtor<CommandInstance> {
  id: string
  category: string
  examples: string[]
  create: (flags, args) => Promise<CommandInstance>
}

export type Validation<UI> = (input: UI) => Promise<boolean>

export interface CommandUX {
  category: string
  function: string
  suffixes?: string[]
  examples: string[]
}

export interface Input<UI, CI> {
  user: UI
  contract: CI | CI[]
}

export const makeCommandId = (category: string, fn: string, suffixes?: string[]): string => {
  const base = `${category}:${fn}`
  return suffixes?.length > 0 ? `${base}:${suffixes.join(':')}` : base
}
