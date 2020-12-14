package node

import (
	"log"
	"net/rpc"
	"sync"
	"time"
)

const numberOfWaitingMessages = 100

// RemoteNode describes remote interface of a gossip node
type RemoteNode struct {
	address            string
	client             *rpc.Client
	waitingMessageChan chan Message
	wg                 sync.WaitGroup
	done               chan struct{}
	errorHandler       func(string, error)
	log                *log.Logger
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
	rn.waitingMessageChan = make(chan Message, numberOfWaitingMessages)
	rn.wg = sync.WaitGroup{}
	rn.done = make(chan struct{}, 1)

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
		rn.client = nil
	}
}

// AttachErrorHandler attaches an error handler to handle rpc method call errors
func (rn *RemoteNode) attachErrorHandler(handler func(string, error)) {
	rn.errorHandler = handler
}

// Send enques a message to send to specific peer
//TODO: How to handle errors
func (rn *RemoteNode) Send(message Message) error {
	/*
		select {
		case rn.waitingMessageChan <- message:
			return nil
		default:
			return errors.New(MessageQueuFullError)
		}
	*/
	// my be it should block
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

			//startTime := time.Now()

			//rn.log.Printf("[RemoteNode-%s]Sending message %s \n", rn.address, m.Base64EncodedHash())
			var response Response
			//err := rn.client.Call("GossipNode.Send", m, &response)

			// sends messages concurrently
			call := rn.client.Go("GossipNode.Send", m, &response, nil)

			if len(m.Payload) < 1000 {
				go rn.checkResultOfAsycCall(call, nil)
			} else {
				startTime := time.Now()
				go rn.checkResultOfAsycCall(call, &startTime)
			}

			//log.Printf("[%s] Elapsed time to send a message with %d bytes is %d \n", rn.address, len(m.Payload), time.Since(startTime).Microseconds())

		}

	}
}

func (rn *RemoteNode) checkResultOfAsycCall(call *rpc.Call, startTime *time.Time) {

	res := <-call.Done

	if res.Error != nil {

		if rn.errorHandler != nil {
			rn.errorHandler(rn.address, res.Error)
			//TODO: breaks the loop,
			//simple error handler could try to reconnect. !!How to handle this case!!!!
			return
		}

		log.Printf("An error occured during sending message to node %s %s \n", rn.address, res.Error)

	}

	if startTime != nil {
		elapsedTime := time.Since(*startTime).Milliseconds()
		m := res.Args.(Message)
		log.Printf("[%s] Message sended in %d ms. Length of the payload is %d \n", rn.address, elapsedTime, len(m.Payload))
	}

}

func (rn *RemoteNode) setLogger(logger *log.Logger) {
	rn.log = logger
}
