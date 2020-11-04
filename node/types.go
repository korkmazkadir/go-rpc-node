package node

import (
	"crypto/sha256"
	"encoding/base64"
)

// Message defines structure of network layer message
type Message struct {
	Layer   MessageLayer
	Tag     string
	Payload []byte
	Forward func()
	//the address of sender 127.0.0.1:3456
	Sender string
}

// Hash return hash256 digest of the message payload
func (m Message) Hash() string {
	return string(m.hash())
}

// Base64EncodedHash returns base64 encoded digest of the message payload
// TODO: this is called in may place and each time it is computed
func (m Message) Base64EncodedHash() string {
	return base64.StdEncoding.EncodeToString(m.hash())
}

func (m Message) hash() []byte {
	h := sha256.New()
	h.Write(m.Payload)
	return h.Sum(nil)
}

// MessageLayer defines layer of a message (NETWORK or APPLICATION)
type MessageLayer int

const (
	// NETWORK layer messages handled by GosipNode
	NETWORK = iota
	// APPLICATION layer messages handled forwarded to registred application by the GosipNode
	APPLICATION = iota
)

// Response is an empty struct to define response of Send rpc call on GosipNode
type Response struct {
}

// Application interface defines the contruct between an application and GossipNode
type Application interface {
	SignalChannel() chan struct{}
	HandleMessage(message Message)
	OutgoingMessageChannel() chan Message
}

// ConnectionRequest is a special NETWORK layer message
// It is used to initiate a bidirectional connection
type ConnectionRequest struct {
	SenderAddress string
}

// MessageQueuFullError sended incase of a message could not put the remote nodes message channel
const MessageQueuFullError = "node message queue is full, could not enqueue the message"
