package node

import (
	"log"
	"net/rpc"
	"time"
)

const numberOfWaitingMessages = 100

// if message payload length bigger than printSendElapsedTimeLimit,
// elapsed time to send the message is logged
// printSendElapsedTimeLimit is in bytes
const bigMessageSize = 25000

// RemoteNode describes remote interface of a gossip node
type RemoteNode struct {
	address string
	client  *rpc.Client
	//to send small messages
	clientSmall *rpc.Client

	waitingMessageChan chan *Message
	done               chan struct{}
	errorHandler       func(string, error)
	log                *log.Logger
	err                error
	rateLimiter        *RateLimiter
}

// NewRemoteNode creates a remote node
func NewRemoteNode(address string) (*RemoteNode, error) {

	//Connects to a remote node and creates a client
	client, err := rpc.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	//Connects to a remote node and creates a client
	clientSmall, err := rpc.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	rn := new(RemoteNode)
	rn.client = client
	rn.clientSmall = clientSmall
	rn.address = address
	rn.waitingMessageChan = make(chan *Message, numberOfWaitingMessages)
	rn.done = make(chan struct{}, 1)
	rn.err = nil

	// Starts a thread to send messages
	// There is only a single thread for a each peer
	go rn.mainLoop()

	return rn, nil
}

//Connect sends a connection request message to the remote node
func (rn *RemoteNode) Connect(nodeAddress string) error {

	connectionRequest := ConnectionRequest{
		SenderAddress: nodeAddress,
	}

	m := Message{
		Layer:   network,
		Tag:     "ConnectionRequest",
		Payload: EncodeToByte(connectionRequest),
	}

	var response Response
	err := rn.client.Call("GossipNode.Send", m, &response)

	return err
}

// Close closes the remote connection to peer.
func (rn *RemoteNode) Close() {
	if rn.client != nil {
		// sends signal to exit from main loop
		rn.done <- struct{}{}
		// closes rpc client
		rn.client.Close()
		// closes channel
		// closing channel does not seems necessary here
		// think about it later
		// close(rn.waitingMessageChan)
	}
}

// Send enques a message to send to specific peer
func (rn *RemoteNode) Send(message *Message) error {
	if rn.err != nil {
		return rn.err
	}

	rn.waitingMessageChan <- message
	return nil
}

func (rn *RemoteNode) mainLoop() {
	for {

		select {

		case <-rn.done:
			log.Printf("Connection to remote node %s  is closed \n", rn.address)
			return

		case m := <-rn.waitingMessageChan:

			//go rn.sendMessage(m)
			rn.sendMessage(m)

		}

	}
}

func (rn *RemoteNode) sendMessage(message *Message) {

	// sending small messsages
	if len(message.Payload) < bigMessageSize {
		rn.sendSmallMessage(message)
		return
	}

	// sending a big message without rate limiting
	if rn.rateLimiter == nil {
		rn.sendBigMessage(message)
		return
	}

	// sending a big message with rate limiting
	// cost of the message equals to cost of the payload
	cost := len(message.Payload)
	rn.rateLimiter.ApplyRateLimit(func() { rn.sendBigMessage(message) }, cost)

}

func (rn *RemoteNode) sendBigMessage(message *Message) {
	startTime := time.Now()
	err := rn.client.Call("GossipNode.Send", *message, nil)
	elapsedTime := time.Since(startTime).Milliseconds()

	if err != nil {
		rn.err = err
		log.Printf("An error occured during sending message to node %s %s. This will close the connection to the remote node! \n", rn.address, err)
		rn.Close()
		return
	}

	log.Printf("[upload-stat]\t%s\t%d\t%d\t%s\n", rn.address, len(message.Payload), elapsedTime, message.Tag)
}

func (rn *RemoteNode) sendSmallMessage(message *Message) {
	err := rn.clientSmall.Call("GossipNode.Send", *message, nil)
	if err != nil {
		rn.err = err
		log.Printf("An error occured during sending message to node %s %s. This will close the connection to the remote node! \n", rn.address, err)
		rn.Close()
	}
}

func (rn *RemoteNode) setLogger(logger *log.Logger) {
	rn.log = logger
}
