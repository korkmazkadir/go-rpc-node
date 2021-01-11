package node

import (
	"log"
	"net/rpc"
	"sync"
	"time"
)

const numberOfWaitingMessages = 100

// if message payload length bigger than printSendElapsedTimeLimit,
// elapsed time to send the message is logged
// printSendElapsedTimeLimit is in bytes
const printSendElapsedTimeLimit = 5000

// RemoteNode describes remote interface of a gossip node
type RemoteNode struct {
	address            string
	client             *rpc.Client
	waitingMessageChan chan *Message
	wg                 sync.WaitGroup
	done               chan struct{}
	errorHandler       func(string, error)
	log                *log.Logger
	err                error
}

// NewRemoteNode creates a remote node
func NewRemoteNode(address string) (*RemoteNode, error) {

	//Connects to a remote node and creates a client
	client, err := rpc.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	rn := new(RemoteNode)
	rn.client = client
	rn.address = address
	rn.waitingMessageChan = make(chan *Message, numberOfWaitingMessages)
	rn.wg = sync.WaitGroup{}
	rn.done = make(chan struct{}, 1)
	rn.err = nil

	// Starts a thread to send messages
	// There is only a single thread for a each peer
	rn.wg.Add(1)
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
		rn.done <- struct{}{}
		rn.client.Close()
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
			rn.wg.Done()
			return

		case m := <-rn.waitingMessageChan:

			startTime := time.Now()
			call := rn.client.Go("GossipNode.Send", *m, nil, nil)
			go rn.checkResultOfAsycCall(call, startTime)

		}

	}
}

func (rn *RemoteNode) checkResultOfAsycCall(call *rpc.Call, startTime time.Time) {

	res := <-call.Done

	if res.Error != nil {
		rn.err = res.Error
		log.Printf("An error occured during sending message to node %s %s. This will close the connection to the remote node! \n", rn.address, res.Error)
		rn.Close()
	}

	m := res.Args.(Message)
	if len(m.Payload) > printSendElapsedTimeLimit {
		elapsedTime := time.Since(startTime).Milliseconds()
		log.Printf("[%s] Message sended in %d ms. Length of the payload is %d \n", rn.address, elapsedTime, len(m.Payload))
	}

}

func (rn *RemoteNode) setLogger(logger *log.Logger) {
	rn.log = logger
}
