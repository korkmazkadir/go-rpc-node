package node

import (
	"log"
	"net"
	"net/rpc"
	"os"
	"sync"
)

const NumberOfBroadcastThread = 5
const MaxProcessedMessageMemorySize = 100

type GossipNode struct {
	App                Application
	peers              []RemoteNode
	broadcastChan      chan Message
	waitChan           chan struct{}
	address            string
	mutex              *sync.Mutex
	forwardMessageChan chan Message
}

// NewGossipNode creates a GossipNode
func NewGossipNode(app Application) *GossipNode {
	node := new(GossipNode)
	node.App = app
	node.peers = make([]RemoteNode, 0)
	node.broadcastChan = make(chan Message, NumberOfBroadcastThread)
	node.waitChan = make(chan struct{}, 1)
	node.mutex = &sync.Mutex{}
	node.forwardMessageChan = make(chan Message, 10)

	return node
}

// Send rpc
func (node *GossipNode) Send(message *Message, reply *Response) error {
	//log.Printf("[GossipNode-%s] tag: %s, payload: %s \n", node.address, message.Tag, message.Payload)

	if message.Layer == NETWORK {

		if message.Tag == "ConnectionRequest" {
			cr := ConnectionRequest{}
			DecodeFromByte(message.Payload, &cr)
			log.Printf("new connection request %+v", cr)
			return node.acceptConnectionRequest(cr)
		}
		log.Printf("unknown message tag %+v", message)
		return nil
	}

	// Here I need the sender nodes information!!!!
	message.Forward = func() {
		node.forwardMessageChan <- *message
	}

	node.App.HandleMessage(*message)
	return nil
}

// It broadcast messages. It sends to all peers, implement except here
func (node *GossipNode) broadcast() {
	outgoingMessageChannel := node.App.OutgoingMessageChannel()

	for {
		select {
		case m := <-outgoingMessageChannel:

			//set sender address of message
			m.Sender = node.address

			log.Println("broadcasts a message")
			for _, peer := range node.peers {
				peer.Send(m)
			}

		case m := <-node.forwardMessageChan:
			node.forward(m, m.Sender)

		}
	}
}

func (node *GossipNode) forward(message Message, exceptNodeAddress string) {

	//sets the address of the sender
	message.Sender = node.address

	log.Println("forwards a message except ", exceptNodeAddress)
	for _, peer := range node.peers {
		if peer.address != exceptNodeAddress {
			peer.Send(message)
		}
	}
}

func (node *GossipNode) listenAndServe(listener *net.TCPListener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			// check if timeout error
			netErr, ok := err.(net.Error)
			if ok && netErr.Timeout() && netErr.Temporary() {
				continue
			} else {
				// do something with bad errors
				log.Printf("connection error: %v", err)
				// end server process, unsucessfully
				os.Exit(1)
			}
		} else {
			// print connection info
			//log.Printf("received rpc message %v %v", reflect.TypeOf(conn), conn)
			// handle listener (client) connections via rpc
			// using a goroutine (to handle more than one connection at a time)
			go rpc.ServeConn(conn)
		}

	}

}

// Start run threads and signals the application. It blocks
// it returns the address of the node
func (node *GossipNode) Start() (string, error) {

	rpc.Register(node)

	// create tcp address
	tcpAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}

	// tcp network listener
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return "", err
	}

	log.Println("server started on ", listener.Addr())

	//signal application
	node.App.SignalChannel() <- struct{}{}

	//starts threads to handle outgoing messages
	go node.broadcast()

	//starts a thread to relpy incomming requests
	go node.listenAndServe(listener)

	//address := fmt.Sprintf("%s:%d", listener.Addr().String())
	node.address = listener.Addr().String()
	return node.address, nil
}

// Wait waits for the node to stop
func (node *GossipNode) Wait() {
	<-node.waitChan
}

// AddPeer adds a peer to the node
func (node *GossipNode) AddPeer(remote RemoteNode) error {

	err := remote.Connect(node.address)
	if err != nil {
		return err
	}

	node.peers = append(node.peers, remote)
	return nil
}

//TODO not thread safe because of append
func (node *GossipNode) acceptConnectionRequest(request ConnectionRequest) error {

	node.mutex.Lock()
	defer node.mutex.Unlock()

	rm, err := NewRemoteNode(request.SenderAddress)
	if err != nil {
		return err
	}

	log.Println("new peer connection added: ", request.SenderAddress)

	node.peers = append(node.peers, *rm)

	return nil
}
