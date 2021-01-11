package node

import (
	"encoding/base32"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"

	"github.com/korkmazkadir/go-rpc-node/filter"
)

const inventoryReadyTag = "INV"
const inventoryRequestTag = "INVR"

// if message payload length is smaller than inventoryMessageLimit,
// it is sended directly. Otherwise an inv message sended
// inventoryMessageLimit is in bytes
const inventoryMessageLimit = 5000

//GossipNode keeps state of a gossip node
type GossipNode struct {
	App                Application
	peerMap            map[string]*RemoteNode
	wg                 sync.WaitGroup
	address            string
	peerMutex          *sync.Mutex
	forwardMessageChan chan Message

	incommingMessageFilter *filter.UniqueMessageFilter

	// messageInventory is not thread safe use with care
	messageInventory FileBackedMessageInventory

	log *log.Logger
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

	// 60 seconds TTL seems reasonable for me
	// I should get this as a parameter
	node.incommingMessageFilter = filter.NewUniqueMessageFilter(120)
	node.messageInventory = NewFileBackedMessageInventory()

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
		} else if message.Tag == inventoryReadyTag {

			added := n.incommingMessageFilter.IfNotContainsAdd(string(message.Payload))
			if added == true {
				n.log.Printf("inventory ready message from %s\n", message.Sender)
				inventoryRequestMessage := n.createInventoryRequestMessage(message)

				n.peerMutex.Lock()
				defer n.peerMutex.Unlock()

				peer := n.peerMap[message.Sender]
				peer.Send(inventoryRequestMessage)
			}

			return nil

		} else if message.Tag == inventoryRequestTag {

			n.log.Printf("inventory request message from %s\n", message.Sender)
			requestedMessageHash := message.Payload
			requestedMessage, err := n.messageInventory.Get(base32.StdEncoding.EncodeToString(requestedMessageHash))
			if err != nil {
				panic(fmt.Errorf("An error occured while getting message from the inventory: %s", err))
			}

			n.peerMutex.Lock()
			defer n.peerMutex.Unlock()

			peer := n.peerMap[message.Sender]
			peer.Send(requestedMessage)

			return nil
		}

		n.log.Printf("Unknown message tag for NETWORK layer message tag: %s \n", message.Tag)
		return nil
	} else if message.Layer == application {

		// Upper layers can call Forward function to forward a message excep the sender
		message.Forward = func() {
			n.forwardMessageChan <- *message
		}

		// Upper layers can call Reply function to reply the sender with a specific message
		message.Reply = func(reply *Message) {
			// Is this correct?
			peerToReply := n.peerMap[message.Sender]
			err := peerToReply.Send(message)
			if err != nil {
				panic(err)
			}
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

	if len(message.Payload) > inventoryMessageLimit {
		// if the length of the nessage payload is bigger than the inventoryMessageLimit
		// an inventory message sended to peers
		n.messageInventory.Put(&message)
		inventoryMessage := n.createInventoryReadyMessage(&message)
		n.sendExceptAllPeers(inventoryMessage, exceptNodeAddress)
	} else {
		// The message sended directly
		n.sendExceptAllPeers(&message, exceptNodeAddress)
	}

}

func (n *GossipNode) sendExceptAllPeers(message *Message, exceptNodeAddress string) {
	n.peerMutex.Lock()
	defer n.peerMutex.Unlock()

	var failedConnections []string

	for address, peer := range n.peerMap {

		if address == exceptNodeAddress {
			continue
		}

		// it is sending an inventory message not the actual message
		err := peer.Send(message)
		if err != nil {
			n.log.Printf("Error occured during sending a message to peer %s. Error: %s\n", address, err)
			failedConnections = append(failedConnections, address)
		}

	}

	for _, failedNodeAddress := range failedConnections {
		delete(n.peerMap, failedNodeAddress)
		log.Printf("WARNING: Removed a failed peer(%s) from peer list. The Number of peers is %d", failedNodeAddress, len(n.peerMap))
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
			go rpc.ServeConn(conn)
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

	previousConnection, isAvailable := n.peerMap[address]
	if isAvailable == true {
		n.log.Printf("Closing the previous connection to %s\n", address)
		previousConnection.Close()
	}
	n.peerMap[address] = peer
	n.log.Printf("New peer added to peer map %s\n", address)
}

func (n *GossipNode) createInventoryReadyMessage(message *Message) *Message {
	inventoryMessage := new(Message)
	inventoryMessage.Sender = n.address
	inventoryMessage.Tag = inventoryReadyTag
	inventoryMessage.Layer = network
	inventoryMessage.Payload = message.hash()

	return inventoryMessage
}

func (n *GossipNode) createInventoryRequestMessage(message *Message) *Message {
	inventoryMessage := new(Message)
	inventoryMessage.Sender = n.address
	inventoryMessage.Tag = inventoryRequestTag
	inventoryMessage.Layer = network
	//inventory message contains only hash of a message as a payload
	inventoryMessage.Payload = message.Payload

	return inventoryMessage
}
