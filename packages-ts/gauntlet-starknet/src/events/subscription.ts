// TODO: Use starknet WS client
type WSClient = any

// TODO: Adapt to Starknet client
export abstract class EventSubscription<Event> {
  abstract parseEvent: (event: any) => Event

  private _wsClient: WSClient

  constructor(readonly client: WSClient) {
    this._wsClient = client
  }

  public start() {
    this._wsClient.start()
  }

  public destroy() {
    this._wsClient.destroy()
  }

  public onEvent(contract: string, eventId: string, callback: (event: Event) => void) {
    this._wsClient.subscribeTx(
      {
        [eventId]: contract,
      },
      (data) => callback(this.parseEvent(data)),
    )
  }
}
