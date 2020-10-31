package node

import (
	"log"
	"net"
	"net/rpc"
	"os"
	"reflect"
)

const NumberOfBroadcastThread = 5

type GossipNode struct {
	App           Application
	peers         []RemoteNode
	broadcastChan chan Message
	waitChan      chan struct{}
	address       string
}

// NewGossipNode creates a GossipNode
func NewGossipNode(app Application) *GossipNode {
	node := new(GossipNode)
	node.App = app
	node.peers = make([]RemoteNode, 0)
	node.broadcastChan = make(chan Message, NumberOfBroadcastThread)
	node.waitChan = make(chan struct{}, 1)

	return node
}

// Send rpc
func (node *GossipNode) Send(message *Message, reply *Response) error {
	log.Printf("[GossipNode-%s] tag: %s, payload: %s \n", node.address, message.Tag, message.Payload)
	node.App.HandleMessage(*message)
	return nil
}

// It broadcast messages. It sends to all peers, implement except here
func (node *GossipNode) broadcast() {
	outgoingMessageChannel := node.App.OutgoingMessageChannel()

	for {
		select {
		case m := <-outgoingMessageChannel:
			for _, peer := range node.peers {
				peer.Send(m)
			}
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
			log.Printf("received rpc message %v %v", reflect.TypeOf(conn), conn)
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
func (node *GossipNode) AddPeer(remote RemoteNode) {
	node.peers = append(node.peers, remote)
}
