package node

import (
	"crypto/sha256"
)

type Message struct {
	Layer   MessageLayer
	Tag     string
	Payload []byte
	Forward func()
	//the address of sender 127.0.0.1:3456
	Sender string
}

func (m Message) Hash() string {
	h := sha256.New()
	h.Write(m.Payload)
	hashValue := string(h.Sum(nil))

	return hashValue
}

type MessageLayer int

const (
	NETWORK     = iota
	APPLICATION = iota
)

type Response struct {
}

type Application interface {
	SignalChannel() chan struct{}
	HandleMessage(message Message)
	OutgoingMessageChannel() chan Message
}

//Connection Request

type ConnectionRequest struct {
	SenderAddress string
}
