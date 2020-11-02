package node

import (
	"log"
	"net/rpc"
)

const NumberOfWaitingMessages = 5

// RemoteNode describes remote interface of a gossip node
type RemoteNode struct {
	address            string
	client             *rpc.Client
	waitingMessageChan chan Message
}

// NewRemoteNode creates a remote node
func NewRemoteNode(address string) (*RemoteNode, error) {
	remoteNode := new(RemoteNode)
	remoteNode.waitingMessageChan = make(chan Message, NumberOfWaitingMessages)

	//Connects to a remote node and creates a client
	remoteNode.address = address
	client, err := rpc.Dial("tcp", remoteNode.address)
	if err != nil {
		return nil, err
	}
	remoteNode.client = client

	// starts a thread to send messages
	// There is only a singl thread for a specific peer
	go remoteNode.mainLoop()

	return remoteNode, nil
}

//Connect sends a connection request message to the remote node
func (remoteNode *RemoteNode) Connect(nodeAddress string) error {

	connectionRequest := ConnectionRequest{
		SenderAddress: nodeAddress,
	}

	m := Message{
		Layer:   NETWORK,
		Tag:     "ConnectionRequest",
		Payload: EncodeToByte(connectionRequest),
	}

	var response Response
	err := remoteNode.client.Call("GossipNode.Send", m, &response)

	return err
}

// Send enques a message to send to specific peer
//TODO: How to handle errors
func (remoteNode *RemoteNode) Send(message Message) {
	remoteNode.waitingMessageChan <- message
}

func (remoteNode *RemoteNode) mainLoop() {
	for {

		select {
		case m := <-remoteNode.waitingMessageChan:

			log.Printf("sending message with tag %s to remote node %s \n", m.Tag, remoteNode.address)
			var response Response
			err := remoteNode.client.Call("GossipNode.Send", m, &response)
			if err != nil {
				log.Printf("an error occured during sending message to node %s %s \n", remoteNode.address, err)
			}

		}

	}
}
