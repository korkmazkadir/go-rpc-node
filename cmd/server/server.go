package main

import (
	"log"

	"../../application"
	"../../node"
)

func main() {

	twoNodes()

	/*
		app := application.NewSimpleApp()
		node := node.NewGossipNode(app)

		app.Start()
		address, err := node.Start()
		if err != nil {
			panic(err)
		}

		log.Printf("node started and listening on %s \n", address)

		node.Wait()
	*/

}

func twoNodes() {

	app1 := application.NewSimpleApp(1)
	node1 := node.NewGossipNode(app1)

	app1.Start()
	address1, err := node1.Start()
	if err != nil {
		panic(err)
	}

	log.Printf("node1 address is %s\n", address1)

	app2 := application.NewSimpleApp(2)
	node2 := node.NewGossipNode(app2)

	app2.Start()
	address2, err := node2.Start()
	if err != nil {
		panic(err)
	}

	remoteNode2, err := node.NewRemoteNode(address2)
	if err != nil {
		panic(err)
	}

	node1.AddPeer(*remoteNode2)

	node1.Wait()
	node2.Wait()

}
