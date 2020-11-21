package node

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"
)

//GossipNode keeps state of a gossip node
type GossipNode struct {
	App                Application
	peerMap            map[string]*RemoteNode
	wg                 sync.WaitGroup
	address            string
	peerMutex          *sync.Mutex
	forwardMessageChan chan Message
	log                *log.Logger
}

// NewGossipNode creates a GossipNode, message buffer size is forward channel buffer size
func NewGossipNode(app Application, messageBufferSize int, logger *log.Logger) *GossipNode {
	node := new(GossipNode)
	node.App = app
	node.peerMap = make(map[string]*RemoteNode)
	node.wg = sync.WaitGroup{}
	node.peerMutex = &sync.Mutex{}
	node.forwardMessageChan = make(chan Message, messageBufferSize)
	node.log = logger

	return node
}

// Send rpc
func (n *GossipNode) Send(message *Message, reply *Response) error {

	if message.Layer == network {

		if message.Tag == "ConnectionRequest" {
			cr := ConnectionRequest{}
			DecodeFromByte(message.Payload, &cr)
			n.log.Printf("New connection request %+v", cr)
			return n.acceptConnectionRequest(cr)
		}

		n.log.Printf("Unknown message tag for NETWORK layer message: %s \n", message.Tag)
		return nil
	} else if message.Layer == application {

		// Upper layers can call Forward function to forward a message excep the sender
		message.Forward = func() {
			n.forwardMessageChan <- *message
		}

		n.App.HandleMessage(*message)
		return nil
	}

	n.log.Printf("Unknown message layer %d\n", message.Layer)

	return nil
}

// It broadcast messages. It sends to all peers, implement except here
func (n *GossipNode) broadcastLoop() {

	outgoingMessageChannel := n.App.OutgoingMessageChannel()

	for {
		select {
		case m := <-outgoingMessageChannel:
			n.forward(m, "")
		case m := <-n.forwardMessageChan:
			n.forward(m, m.Sender)
		}
	}
}

func (n *GossipNode) forward(message Message, exceptNodeAddress string) {

	//sets the address of the sender
	message.Sender = n.address

	if exceptNodeAddress == "" {
		n.log.Printf("Broadcasts the message %s \n", message.Base64EncodedHash())
	} else {
		n.log.Printf("Forwards the message %s except %s \n", message.Base64EncodedHash(), exceptNodeAddress)
	}

	n.peerMutex.Lock()
	defer n.peerMutex.Unlock()

	for address, peer := range n.peerMap {

		if address == exceptNodeAddress {
			continue
		}

		err := peer.Send(message)
		if err != nil {
			n.log.Printf("Error occured during sending a message to peer %s. Error: %s\n", address, err)
		}

	}

}

func (n *GossipNode) listenAndServeLoop(listener *net.TCPListener) {

	for {
		conn, err := listener.Accept()
		if err != nil {
			// check if timeout error
			netErr, ok := err.(net.Error)
			if ok && netErr.Timeout() && netErr.Temporary() {
				continue
			} else {
				// do something with bad errors
				log.Printf("GossipNode Connection error: %v", err)
				// end server process, unsucessfully
				panic(netErr)
			}
		} else {

			// using a goroutine (to handle more than one connection at a time)
			//go rpc.ServeConn(conn)
			rpc.ServeConn(conn)
		}

	}

}

// Start run threads and signals the application. It blocks
// it returns the address of the node
func (n *GossipNode) Start(hostname string) (string, error) {

	//Registrers only send method
	rpc.Register(n)

	// create tcp address
	tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:0", hostname))
	if err != nil {
		return "", err
	}

	// tcp network listener
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return "", err
	}

	n.log.Println("Server started on ", listener.Addr())

	//signal application
	n.App.SignalChannel() <- struct{}{}

	//starts threads to handle outgoing messages
	n.wg.Add(1)
	go n.broadcastLoop()

	//starts a thread to relpy incomming requests
	n.wg.Add(1)
	go n.listenAndServeLoop(listener)

	//address := fmt.Sprintf("%s:%d", listener.Addr().String())
	n.address = listener.Addr().String()
	return n.address, nil
}

// Wait waits for the node to stop
func (n *GossipNode) Wait() {
	n.wg.Wait()
}

// AddPeer adds a peer to the node
func (n *GossipNode) AddPeer(remote *RemoteNode) error {

	err := remote.Connect(n.address)
	if err != nil {
		return err
	}

	remote.setLogger(n.log)

	n.addPeer(remote)
	return nil
}

func (n *GossipNode) acceptConnectionRequest(request ConnectionRequest) error {

	rm, err := NewRemoteNode(request.SenderAddress)
	if err != nil {
		return err
	}

	n.log.Printf("New peer connection request accepted from %s \n", request.SenderAddress)
	n.addPeer(rm)

	return nil
}

func (n *GossipNode) addPeer(peer *RemoteNode) {
	address := peer.address

	n.peerMutex.Lock()
	defer n.peerMutex.Unlock()

	//Attaches to simple error handler to handle rpc errors
	peer.attachErrorHandler(n.simpleErrorHandler)

	previousConnection, isAvailable := n.peerMap[address]
	if isAvailable == true {
		n.log.Printf("Closing the previous connection to %s\n", address)
		previousConnection.Close()
	}
	n.peerMap[address] = peer
	n.log.Printf("New peer added to peer map %s\n", address)
}

func (n *GossipNode) simpleErrorHandler(nodeAddress string, err error) {

	n.peerMutex.Lock()
	defer n.peerMutex.Unlock()

	n.log.Printf("Error occured during sending message to node %s.\n", err)
	n.log.Printf("Connection to %s is shut down.\n", nodeAddress)

	previousConnection, isAvailable := n.peerMap[nodeAddress]
	if isAvailable == true {
		log.Printf("Closing connection to %s because of a send error %s \n", nodeAddress, err)
		previousConnection.Close()
		delete(n.peerMap, nodeAddress)
	}

}
