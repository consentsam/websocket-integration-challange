package clients

// DeltaClient defines the methods used by WebsocketHandler to interact with a data source like Delta Exchange.
type DeltaClient interface {
	Connect() error
	Close() error
	Subscribe(channel string, productIDs []string) error
	Unsubscribe(channel string) error
	RegisterHandler(channel string, handler MessageHandler)
	GetConnectionStatus() map[string]interface{}
	IsConnected() bool
}
