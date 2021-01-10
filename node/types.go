package node

import (
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
)

// MessageLayer defines layer of a message (NETWORK or APPLICATION)
type messageLayer int

const (
	// network layer messages handled by GosipNode
	network = iota
	// application layer messages handled forwarded to registred application by the GosipNode
	application = iota
)

// Message defines structure of network layer message
type Message struct {
	// The address of sender 127.0.0.1:3456
	Sender string
	// Tag defines the type of a message. It is used to convert byte array to an object on the receiver side.
	Tag string
	// Forward is a callback to forward message accordıng to decision of the application
	Forward func() `json:"-"`

	//Reply replies the message with a given message
	Reply func(*Message) `json:"-"`

	// Layer defines the layer of message
	Layer messageLayer
	// Payload keeps the encoded message content
	Payload []byte
}

// NewMessage creates a network message
func NewMessage(tag string, payload []byte) Message {
	message := Message{
		Tag:     tag,
		Payload: payload,
		Layer:   application,
	}
	return message
}

// Hash return hash256 digest of the message payload
func (m Message) Hash() string {
	return string(m.hash())
}

// Base64EncodedHash returns base64 encoded digest of the message payload
func (m Message) Base64EncodedHash() string {
	return base64.StdEncoding.EncodeToString(m.hash())
}

// Base32EncodedHash returns base32 encoded digest of the message payload
func (m Message) Base32EncodedHash() string {
	return base32.StdEncoding.EncodeToString(m.hash())
}

func (m Message) hash() []byte {
	h := sha256.New()
	h.Write(m.Payload)
	return h.Sum(nil)
}

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
