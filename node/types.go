package node

type Message struct {
	Tag     string
	Payload string
}

type Response struct {
}

type Application interface {
	SignalChannel() chan struct{}
	HandleMessage(message Message)
	OutgoingMessageChannel() chan Message
}
