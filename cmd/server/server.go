package main

import (
	"flag"
	"log"

	"../../application"
	"../../node"
)

func main() {

	peerAddress := flag.String("peer", "", "adds a peer by address")

	flag.Parse()

	app := application.NewSimpleApp(1)
	n := node.NewGossipNode(app)
	app.Start()
	address, err := n.Start()
	if err != nil {
		panic(err)
	}

	log.Println(*peerAddress)
	if *peerAddress != "" {
		log.Println("adding a peer")
		peer, err := node.NewRemoteNode(*peerAddress)
		if err != nil {
			panic(err)
		}
		n.AddPeer(*peer)
	}

	log.Println("node started ", address)

	n.Wait()
}
