package main

import (
	"flag"
	"log"
	"strings"

	"../../application"
	"../../node"
)

func main() {

	peerAddresses := flag.String("peers", "", "semicolon seperated list peer addresses")

	flag.Parse()

	app := application.NewSimpleApp(1)
	n := node.NewGossipNode(app)
	app.Start()
	address, err := n.Start()
	if err != nil {
		panic(err)
	}

	log.Println(*peerAddresses)
	if *peerAddresses != "" {

		tokens := strings.Split(*peerAddresses, ";")

		for _, address := range tokens {
			connectToPeer(address, n)
		}

	}

	log.Println("node started ", address)

	n.Wait()
}

func connectToPeer(peerAddress string, n *node.GossipNode) {

	peer, err := node.NewRemoteNode(peerAddress)
	if err != nil {
		panic(err)
	}

	err = n.AddPeer(*peer)
	if err != nil {
		panic(err)
	}

}
