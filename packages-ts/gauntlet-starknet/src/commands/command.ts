export interface CommandCtor<CommandInstance> {
  id: string
  category: string
  examples: string[]
  create: (flags, args) => Promise<CommandInstance>
}

export type Validation<UI> = (input: UI) => Promise<boolean>
